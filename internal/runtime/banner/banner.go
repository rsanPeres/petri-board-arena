package banner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Info struct {
	AppName string

	Env           string
	Port          string
	WriteDBURL    string
	ReadDBURL     string
	GitCommit     string
	BuildTime     string
	SchemaGlob    string
	MigrationsDir string
	NowUTC        time.Time
}

func Print(i Info) {
	if i.NowUTC.IsZero() {
		i.NowUTC = time.Now().UTC()
	}
	if i.AppName == "" {
		i.AppName = "petri-arena"
	}
	if i.MigrationsDir == "" {
		i.MigrationsDir = "migrations"
	}

	writeUp, writeDown := countMigrations(filepath.Join(i.MigrationsDir, "write"))
	readUp, readDown := countMigrations(filepath.Join(i.MigrationsDir, "read"))

	fmt.Println(art(i.AppName))
	fmt.Printf("UTC: %s\n", i.NowUTC.Format(time.RFC3339))
	fmt.Printf("Go:  %s (%s/%s)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if i.Env != "" {
		fmt.Printf("Env: %s\n", i.Env)
	}

	if i.Port != "" {
		fmt.Printf("Port: %s\n", i.Port)
	}

	if i.GitCommit != "" || i.BuildTime != "" {
		bt := i.BuildTime
		if bt == "" {
			bt = "unknown"
		}
		gc := i.GitCommit
		if gc == "" {
			gc = "unknown"
		}
		fmt.Printf("Build: commit=%s time=%s\n", gc, bt)
	}

	if i.WriteDBURL != "" {
		fmt.Printf("WriteDB: %s\n", redactDSN(i.WriteDBURL))
	}
	if i.ReadDBURL != "" {
		fmt.Printf("ReadDB:  %s\n", redactDSN(i.ReadDBURL))
	}

	fmt.Printf("Migrations: write(up=%d, down=%d) read(up=%d, down=%d)\n",
		writeUp, writeDown, readUp, readDown,
	)

	if files := listMigrationFiles(filepath.Join(i.MigrationsDir, "write")); len(files) > 0 {
		fmt.Printf("Write migrations: %s\n", strings.Join(files, ", "))
	}
	if files := listMigrationFiles(filepath.Join(i.MigrationsDir, "read")); len(files) > 0 {
		fmt.Printf("Read migrations:  %s\n", strings.Join(files, ", "))
	}

	fmt.Println(strings.Repeat("-", 72))
}

func countMigrations(dir string) (up int, down int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			up++
		} else if strings.HasSuffix(name, ".down.sql") {
			down++
		}
	}
	return up, down
}

func listMigrationFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".up.sql") {
			files = append(files, strings.TrimSuffix(n, ".up.sql"))
		}
	}
	sort.Strings(files)

	if len(files) > 12 {
		return append(files[:12], fmt.Sprintf("...(+%d)", len(files)-12))
	}
	return files
}

func redactDSN(dsn string) string {

	i := strings.Index(dsn, "://")
	if i == -1 {
		return dsn
	}
	rest := dsn[i+3:]
	at := strings.Index(rest, "@")
	colon := strings.Index(rest, ":")
	if colon == -1 || at == -1 || colon > at {
		return dsn
	}
	user := rest[:colon]
	return dsn[:i+3] + user + ":***@" + rest[at+1:]
}

func art(app string) string {

	return fmt.Sprintf(`
██████╗ ███████╗████████╗██████╗ ██╗       █████╗ ██████╗ ███████╗███╗   ██╗ █████╗
██╔══██╗██╔════╝╚══██╔══╝██╔══██╗██║      ██╔══██╗██╔══██╗██╔════╝████╗  ██║██╔══██╗
██████╔╝█████╗     ██║   ██████╔╝██║      ███████║██████╔╝█████╗  ██╔██╗ ██║███████║
██╔═══╝ ██╔══╝     ██║   ██╔══██╗██║      ██╔══██║██╔══██╗██╔══╝  ██║╚██╗██║██╔══██║
██║     ███████╗   ██║   ██║  ██║██║      ██║  ██║██║  ██║███████╗██║ ╚████║██║  ██║
╚═╝     ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝      ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═══╝╚═╝  ╚═╝

%s
`, app)
}
