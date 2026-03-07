package informer

import (
	"context"
	"fmt"
	"time"

	"aimtp/internal/aimtp_controller/informer/reconciler"
	"aimtp/internal/aimtp_controller/store"
	"aimtp/internal/pkg/k8s"
	"aimtp/internal/pkg/log"

	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	wfinformers "github.com/argoproj/argo-workflows/v3/pkg/client/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	volcanov1 "volcano.sh/apis/pkg/apis/batch/v1alpha1"
	volcanoclientset "volcano.sh/apis/pkg/client/clientset/versioned"
	volcanoinformers "volcano.sh/apis/pkg/client/informers/externalversions"
)

const (
	// ResyncPeriod 是 Informer 的重新同步周期
	ResyncPeriod = 10 * time.Hour
	// WorkerCount 是每个 Controller 的 Worker 数量
	WorkerCount = 2

	// EventKindVJ 是 Volcano Job 的 Kind
	EventKindVJ = "Job"
)

// Informer 是 aimtp-controller 的核心控制器
// 它负责监听 K8s 资源变化，并将其同步到数据库
type Informer struct {
	cluster  string
	jobStore store.JobStore
	restCfg  *rest.Config

	// Clients
	k8sClient     kubernetes.Interface
	argoClient    wfclientset.Interface
	volcanoClient volcanoclientset.Interface

	// Informer Factories
	k8sFactory     informers.SharedInformerFactory
	argoFactory    wfinformers.SharedInformerFactory
	volcanoFactory volcanoinformers.SharedInformerFactory

	// Informers
	podInformer      cache.SharedIndexInformer
	workflowInformer cache.SharedIndexInformer
	vjInformer       cache.SharedIndexInformer

	// WorkQueues
	podQueue      workqueue.TypedRateLimitingInterface[string]
	workflowQueue workqueue.TypedRateLimitingInterface[string]
	vjQueue       workqueue.TypedRateLimitingInterface[string]

	// Reconcilers
	jobReconciler *reconciler.JobReconciler
	podReconciler *reconciler.PodReconciler
}

// New 创建一个新的 Informer
func New(
	cluster string,
	kubeconfig *rest.Config,
	jobStore store.JobStore,
	podStore store.PodStore,
) (*Informer, error) {
	// 1. 初始化 Clients
	k8sClient, err := k8s.NewKubeClient(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("create k8s client: %w", err)
	}

	argoClient, err := k8s.NewArgoClient(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("create argo client: %w", err)
	}

	volcanoClient, err := k8s.NewVolcanoClient(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("create volcano client: %w", err)
	}

	// 2. 初始化 Informer Factories
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, ResyncPeriod)
	argoFactory := wfinformers.NewSharedInformerFactory(argoClient, ResyncPeriod)
	volcanoFactory := volcanoinformers.NewSharedInformerFactory(volcanoClient, ResyncPeriod)

	// 3. 获取具体资源的 Informer
	podInformer := k8sFactory.Core().V1().Pods().Informer()
	workflowInformer := argoFactory.Argoproj().V1alpha1().Workflows().Informer()
	vjInformer := volcanoFactory.Batch().V1alpha1().Jobs().Informer()

	c := &Informer{
		cluster:  cluster,
		jobStore: jobStore,
		restCfg:  kubeconfig,

		k8sClient:     k8sClient,
		argoClient:    argoClient,
		volcanoClient: volcanoClient,

		k8sFactory:     k8sFactory,
		argoFactory:    argoFactory,
		volcanoFactory: volcanoFactory,

		podInformer:      podInformer,
		workflowInformer: workflowInformer,
		vjInformer:       vjInformer,

		podQueue:      workqueue.NewTypedRateLimitingQueueWithConfig[string](workqueue.DefaultTypedControllerRateLimiter[string](), workqueue.TypedRateLimitingQueueConfig[string]{Name: "Pods"}),
		workflowQueue: workqueue.NewTypedRateLimitingQueueWithConfig[string](workqueue.DefaultTypedControllerRateLimiter[string](), workqueue.TypedRateLimitingQueueConfig[string]{Name: "Workflows"}),
		vjQueue:       workqueue.NewTypedRateLimitingQueueWithConfig[string](workqueue.DefaultTypedControllerRateLimiter[string](), workqueue.TypedRateLimitingQueueConfig[string]{Name: "VolcanoJobs"}),

		jobReconciler: reconciler.NewJobReconciler(cluster, jobStore, podStore),
		podReconciler: reconciler.NewPodReconciler(cluster, podStore),
	}

	// 4. 注册 Event Handlers
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueue(c.podQueue, obj)
		},
		UpdateFunc: func(old, new interface{}) {
			c.enqueue(c.podQueue, new)
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueue(c.podQueue, obj)
		},
	})

	workflowInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueue(c.workflowQueue, obj)
		},
		UpdateFunc: func(old, new interface{}) {
			c.enqueue(c.workflowQueue, new)
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueue(c.workflowQueue, obj)
		},
	})

	vjInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueue(c.vjQueue, obj)
		},
		UpdateFunc: func(old, new interface{}) {
			c.enqueue(c.vjQueue, new)
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueue(c.vjQueue, obj)
		},
	})

	return c, nil
}

// enqueue 将对象放入 WorkQueue
func (c *Informer) enqueue(queue workqueue.TypedRateLimitingInterface[string], obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		return
	}
	queue.Add(key)
}

// Run 启动 Controller
func (c *Informer) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.podQueue.ShutDown()
	defer c.workflowQueue.ShutDown()
	defer c.vjQueue.ShutDown()

	log.Infow("Starting Informer Controller", "cluster", c.cluster)

	// 启动 Informer Factories
	c.k8sFactory.Start(stopCh)
	c.argoFactory.Start(stopCh)
	c.volcanoFactory.Start(stopCh)

	// 等待缓存同步
	log.Infow("Waiting for informer caches to sync")
	if !cache.WaitForCacheSync(stopCh,
		c.podInformer.HasSynced,
		c.workflowInformer.HasSynced,
		c.vjInformer.HasSynced,
	) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Infow("Starting workers", "count", WorkerCount)
	for i := 0; i < WorkerCount; i++ {
		go wait.Until(c.runPodWorker, time.Second, stopCh)
		go wait.Until(c.runVJWorker, time.Second, stopCh)
		// go wait.Until(c.runWorkflowWorker, time.Second, stopCh) // TODO: Argo Reconciler
	}

	log.Infow("Informer Controller started")
	<-stopCh
	log.Infow("Shutting down Informer Controller")

	return nil
}

// runPodWorker 处理 Pod 队列
func (c *Informer) runPodWorker() {
	for c.processNextItem(c.podQueue, c.syncPod) {
	}
}

// runVJWorker 处理 Volcano Job 队列
func (c *Informer) runVJWorker() {
	for c.processNextItem(c.vjQueue, c.syncVolcanoJob) {
	}
}

// processNextItem 通用处理逻辑
func (c *Informer) processNextItem(
	queue workqueue.TypedRateLimitingInterface[string],
	syncFunc func(string) error,
) bool {
	key, quit := queue.Get()
	if quit {
		return false
	}
	defer queue.Done(key)

	err := syncFunc(key)
	if err == nil {
		queue.Forget(key)
		return true
	}

	if queue.NumRequeues(key) < 5 {
		log.Warnw("Error syncing item, requeueing", "key", key, "err", err)
		queue.AddRateLimited(key)
	} else {
		log.Errorw("Dropping item after max retries", "key", key, "err", err)
		queue.Forget(key)
	}
	return true
}

// syncPod 同步 Pod 状态
func (c *Informer) syncPod(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// 从 Lister/Indexer 获取对象
	obj, exists, err := c.podInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		// Pod 被删除，我们可能需要更新状态为 Deleted，或者忽略
		// 目前 PodReconciler 主要处理 Update/Create，删除逻辑可以在这里补充
		log.Debugw("Pod deleted", "namespace", namespace, "name", name)
		if err := c.podReconciler.Delete(context.Background(), namespace, name); err != nil {
			log.Errorw("Failed to delete pod from DB", "name", name, "err", err)
			return err
		}
		return nil
	}

	pod := obj.(*corev1.Pod)
	return c.podReconciler.Reconcile(context.Background(), pod)
}

// syncVolcanoJob 同步 Volcano Job 状态
func (c *Informer) syncVolcanoJob(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.vjInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		log.Debugw("Volcano Job deleted", "namespace", namespace, "name", name)
		// 调用 Delete 方法处理删除逻辑
		if err := c.jobReconciler.Delete(context.Background(), namespace, name); err != nil {
			log.Errorw("Failed to delete job from DB", "name", name, "err", err)
			return err
		}
		return nil
	}

	job := obj.(*volcanov1.Job)
	return c.jobReconciler.Reconcile(context.Background(), job)
}
