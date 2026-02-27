package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	backupcmd "github.com/graperegistry/grape/cmd/backup"
	"github.com/graperegistry/grape/internal/config"
	"github.com/graperegistry/grape/internal/db"
	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/server"
	"github.com/graperegistry/grape/internal/server/handler"
)

var (
	configPath string
	version    = "0.1.0"
)

func init() {
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&configPath, "c", "", "Path to config file (shorthand)")
}

func printUsage() {
	fmt.Println("Grape - Lightweight private npm registry")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  grape [options]              Start the server")
	fmt.Println("  grape backup [options]       Create a backup")
	fmt.Println("  grape restore [options]      Restore from backup")
	fmt.Println("  grape list [options]         List backup contents")
	fmt.Println()
	fmt.Println("Server Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Backup Options:")
	fmt.Println("  --output, -o    Output file path (default: grape-backup-YYYYMMDD-HHMMSS.tar.gz)")
	fmt.Println()
	fmt.Println("Restore Options:")
	fmt.Println("  --input, -i     Input backup file (required)")
	fmt.Println("  --force, -f     Force overwrite existing data")
	fmt.Println()
	fmt.Println("List Options:")
	fmt.Println("  --input, -i     Input backup file (required)")
}

func main() {
	// 检查是否有子命令
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		// 子命令模式
		switch os.Args[1] {
		case "backup":
			runBackupCommand(os.Args[2:])
			return
		case "restore":
			runRestoreCommand(os.Args[2:])
			return
		case "list":
			runListCommand(os.Args[2:])
			return
		case "help", "--help", "-h":
			printUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	// 服务器模式
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.Log.Level); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 初始化数据库
	if err := db.Init(&db.Config{
		Type: cfg.Database.Type,
		DSN:  cfg.Database.DSN,
	}); err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 自动迁移数据库表
	if err := db.Migrate(
		&db.User{}, &db.Package{}, &db.PackageVersion{}, &db.Webhook{},
		&db.AuditLog{}, &db.Token{}, &db.PackageOwner{},
		&db.PackageGCMetadata{}, &db.OrphanedFile{}, &db.PackageDeprecation{},
	); err != nil {
		logger.Fatalf("Failed to migrate database: %v", err)
	}

	// 运行 SQL 迁移
	if err := db.RunMigrations(db.DB); err != nil {
		logger.Warnf("Failed to run SQL migrations: %v", err)
	}

	logger.Info("✅ Database initialized")

	srv := server.New(cfg, version)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 停止登录限流器
	handler.StopLoginLimiter()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	logger.Info("👋 Grape stopped")
}

func runBackupCommand(args []string) {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	output := fs.String("output", "", "Output file path")
	fs.StringVar(output, "o", "", "Output file path (shorthand)")

	fs.Usage = func() {
		fmt.Println("Usage: grape backup [options]")
		fmt.Println()
		fmt.Println("Create a backup of Grape data")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	cmd := &backupcmd.BackupCommand{
		Output: *output,
	}

	if err := logger.Init("info"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		logger.Fatalf("Backup failed: %v", err)
	}
}

func runRestoreCommand(args []string) {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	input := fs.String("input", "", "Input backup file")
	force := fs.Bool("force", false, "Force overwrite")
	fs.StringVar(input, "i", "", "Input backup file (shorthand)")
	fs.BoolVar(force, "f", false, "Force overwrite (shorthand)")

	fs.Usage = func() {
		fmt.Println("Usage: grape restore [options]")
		fmt.Println()
		fmt.Println("Restore Grape data from backup")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if *input == "" {
		fmt.Fprintf(os.Stderr, "Error: input file required (--input)\n\n")
		fs.Usage()
		os.Exit(1)
	}

	cmd := &backupcmd.RestoreCommand{
		Input: *input,
		Force: *force,
	}

	if err := logger.Init("info"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		logger.Fatalf("Restore failed: %v", err)
	}
}

func runListCommand(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	input := fs.String("input", "", "Input backup file")
	fs.StringVar(input, "i", "", "Input backup file (shorthand)")

	fs.Usage = func() {
		fmt.Println("Usage: grape list [options]")
		fmt.Println()
		fmt.Println("List contents of a backup file")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if *input == "" {
		fmt.Fprintf(os.Stderr, "Error: input file required (--input)\n\n")
		fs.Usage()
		os.Exit(1)
	}

	cmd := &backupcmd.BackupCommand{
		List:  true,
		Input: *input,
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}