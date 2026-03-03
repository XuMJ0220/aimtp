package main

import (
	"aimtp/pkg/db"
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/pflag"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// 帮助信息文本.
const helpText = `Usage: main [flags] arg [arg...]

This is a pflag example.

Flags:
`

type GenerateConfig struct {
	ModelPackagePath string
	GenerateFunc     func(g *gen.Generator)
}

// 预定义的生成配置.
var generateConfigs = map[string]GenerateConfig{
	"aimpt_user":   {ModelPackagePath: "../../internal/apiserver/model", GenerateFunc: GenerateModels},
	"aimpt_server": {ModelPackagePath: "../../internal/aimtp_server/model", GenerateFunc: GenerateAIMTPServerModels},
}

var (
	addr       = pflag.StringP("addr", "a", "127.0.0.1:3306", "MySQL host address.")
	username   = pflag.StringP("username", "u", "aimtp", "Username to connect to the database.")
	password   = pflag.StringP("password", "p", "aimtp1234", "Password to use when connecting to the database.")
	database   = pflag.StringP("db", "d", "aimtp", "Database name to connect to.")
	modelPath  = pflag.String("model-pkg-path", "", "Generated model code's package name.")
	components = pflag.StringSlice("component", []string{"aimpt_user", "aimpt_server"}, "Generated model code's for specified component.")
	help       = pflag.BoolP("help", "h", false, "Show this help message.")
)

func main() {
	// 设置自定义的使用说明函数
	pflag.Usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults() // 会把可以设置的标志打印出来
	}

	pflag.Parse()

	// 如果设置了帮助标志，则显示帮助信息并退出
	if *help {
		pflag.Usage()
		return
	}

	// 初始化数据库连接
	dbInstance, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 处理组件并生成代码
	for _, component := range *components {
		processComponent(component, dbInstance)
	}
}

// initializeDatabase 创建并返回一个数据库连接.
func initializeDatabase() (*gorm.DB, error) {
	dbOptions := &db.MySQLOptions{
		Addr:     *addr,
		Username: *username,
		Password: *password,
		Database: *database,
	}

	// 创建并返回数据库连接
	return db.NewMySQL(dbOptions)
}

// processComponent 处理单个组件以生成代码.
func processComponent(component string, dbInstance *gorm.DB) {
	config, ok := generateConfigs[component]
	if !ok {
		log.Printf("Component '%s' not found in configuration. Skipping.", component)
		return
	}

	// 解析模型包路径
	modelPkgPath := resolveModelPackagePath(config.ModelPackagePath)

	// 创建生成器实例
	generator := createGenerator(modelPkgPath)
	generator.UseDB(dbInstance)

	// 应用自定义生成器选项
	applyGeneratorOptions(generator)

	// 使用指定的函数生成模型
	config.GenerateFunc(generator)

	// 执行代码生成
	generator.Execute()
}

// resolveModelPackagePath 确定模型生成的存放路径.
func resolveModelPackagePath(defaultPath string) string {
	if *modelPath != "" {
		return *modelPath
	}
	absPath, err := filepath.Abs(defaultPath) // 将相对路径转为绝对路径，保证生成器无论在哪里运行，都能准确找到目标位置
	if err != nil {
		log.Printf("Error resolving path: %v", err)
		return defaultPath
	}
	return absPath
}

// createGenerator 初始化并返回一个新的生成器实例.
func createGenerator(packagePath string) *gen.Generator {
	return gen.NewGenerator(gen.Config{
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		ModelPkgPath:      packagePath,
		WithUnitTest:      true,
		FieldNullable:     true,  // 对于数据库中可空的字段，使用指针类型，以此便可以区分零值和空。
		FieldSignable:     false, // 禁用无符号属性以提高兼容性。
		FieldWithIndexTag: false, // 不包含 GORM 的索引标签。
		FieldWithTypeTag:  false, // 不包含 GORM 的类型标签。
	})
}

// applyGeneratorOptions 设置自定义生成器选项.
func applyGeneratorOptions(g *gen.Generator) {
	// 为特定字段自定义 GORM 标签
	// 每个生成的 go 结构体都必须要有
	g.WithOpts(
		gen.FieldGORMTag("createdAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
		gen.FieldGORMTag("updatedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
	)
}

// GenerateModels 生成模型.
func GenerateModels(g *gen.Generator) {
	g.GenerateModelAs(
		"user",                         // 自动生成 go model 所依赖的表明
		"UserM",                        // 生成的 go model 结构体名
		gen.FieldIgnore("placeholder"), // go 结构体中不要生成 placeholder 字段
		gen.FieldGORMTag("username", func(tag field.GormTag) field.GormTag { // 添加标签 uniqueIndex:idx_user_username
			tag.Set("uniqueIndex", "idx_user_username")
			return tag
		}),
		gen.FieldGORMTag("userID", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_user_userID")
			return tag
		}),
		gen.FieldGORMTag("phone", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_user_phone")
			return tag
		}),
	)
	g.GenerateModelAs(
		"casbin_rule",
		"CasbinRuleM",
		gen.FieldRename("ptype", "PType"), // 把表中 ptype 原本应该自动生成为 go 结构体字段 的 Ptype 改成 PType
		gen.FieldIgnore("placeholder"),
	)
}

func GenerateAIMTPServerModels(g *gen.Generator) {
	g.GenerateModelAs(
		"dag_status_summary",
		"DagStatusSummaryM",
		gen.FieldIgnore("placeholder"),
		gen.FieldGORMTag("dagName", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_dag_status_summary_dag_name")
			return tag
		}),
	)
}
