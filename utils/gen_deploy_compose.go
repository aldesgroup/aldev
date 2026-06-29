package utils

import (
	"fmt"
	"os"
	"path"

	"github.com/aldesgroup/aldev/templates"
	core "github.com/aldesgroup/corego"
)

// ----------------------------------------------------------------------------
// Building a config file for the `podman compose up` command run for local dev
// ----------------------------------------------------------------------------

// the local compose file is complicated in case we have a DB - which should happen quite often. And the content depends on the DB type, obviously.
// We want the "podman compose up" command to help us have a 100% working local deployment.
func generateLocalComposeFile(localDir string, resolvedLocalAPIConfig map[string]interface{}) {
	// checking the DB servers config in the resolved local API config
	dbServerConfigsObj := core.GetValueFromMap(resolvedLocalAPIConfig, "base", "dbservers")

	// no database, this is straightforward
	if dbServerConfigsObj == nil {
		// no DB servers configured, so we just generate a simple compose file with the API and the Nginx
		EnsureFileFromTemplate(path.Join(localDir, "compose.yaml"), templates.LocalCOMPOSE, "", "", "", "")
		return
	}

	// checking the DB servers config in the resolved local API config
	_, ok := dbServerConfigsObj.(map[string]interface{})
	core.PanicMsgIf(!ok, "The 'dbservers' config is not a map[string]interface{} as expected")

	// checking each DB config
	dbConfigs := make(map[string]map[string]interface{})
	for serverName, dbServerConfigObj := range dbServerConfigsObj.(map[string]interface{}) {
		dbConfig, ok := dbServerConfigObj.(map[string]interface{})
		core.PanicMsgIf(!ok, "The 'dbservers.%s' config is not a map[string]interface{} as expected", serverName)
		dbConfigs[serverName] = dbConfig
	}

	// generating
	EnsureFileFromTemplate(path.Join(localDir, "compose.yaml"), templates.LocalCOMPOSE,
		getBeforeApiPart(dbConfigs), getApiRunPart(dbConfigs), getApiDependPart(dbConfigs), getVolumesPart(dbConfigs))
}

// ----------------------------------------------------------------------------
// Filling the placeholders in the compose template file
// ----------------------------------------------------------------------------

// building a part of the compose file
func getBeforeApiPart(dbConfigs map[string]map[string]interface{}) string {
	// for now, we always need to do this
	cleanJob := `    # Removes the migration sentinel file before anything else runs
    {{.AppNameShort}}_clean:
        image: {{.API.Build.RunImage}}
        volumes:
            - ../../tmp:/api/tmp:z
        command:
            ["sh", "-c", "rm -f /api/tmp/.db_init_done; echo done cleaning"]
        restart: "no"
`

	dbServices := getDbServicesPart(dbConfigs)

	initJob := fmt.Sprintf(`
    # Runs the API in migration mode to set up the database schema and initial data
    {{.AppNameShort}}_init:
        image: {{.API.Build.RunImage}}
        # Sets the starting directory inside the container for any relative paths in your code
        working_dir: /api
        volumes:
            # Mounts the host's bin folder to the container.
            # The ':z' tells Podman to relabel the files for SELinux (essential on Fedora/RHEL)
            - ../../{{.ResolvedBinDir}}:/api/bin:z
            - ../../{{.API.Build.SrcDir}}:/api/src:z
            - ../../tmp:/api/tmp:z
        command:
            - sh
            - -c
            - |
                ./bin/{{.AppNameKebab}}-api -config src/conf-local.yaml -migrate && touch tmp/.db_init_done
        depends_on:%s
        restart: "no"
`, getDependOnDbServicesPart(dbConfigs))

	return cleanJob + dbServices + initJob
}

// building a part of the compose file
func getApiRunPart(_ map[string]map[string]interface{}) string {
	return `
                echo "Waiting for migration to complete successfully..."
                retries=0
                max_retries=20
                until [ -f tmp/.db_init_done ]; do
                    printf 'retry=%s\n' "$$retries"
                    if [ "$$retries" -ge "$$max_retries" ]; then
                        echo "Migration sentinel not detected after $${max_retries} retries"
                        exit 1
                    fi
                    sleep 0.5
                    retries=$$((retries + 1))
                    echo "waiting... ($${retries}/$${max_retries})"
                done
                echo "Migration done, starting API"`
}

// building a part of the compose file
func getApiDependPart(dbConfigs map[string]map[string]interface{}) string {
	return fmt.Sprintf(`
        depends_on:%s`, getDependOnDbServicesPart(dbConfigs))
}

// building a part of the compose file
func getVolumesPart(dbConfigs map[string]map[string]interface{}) string {
	if hasPostgreSQLDb(dbConfigs) {
		// let's create a volume for the PostgreSQL database, so that the data is persisted across container restarts
		localEnvCtx := NewBaseContext().WithStdErrWriter(os.Stdout).WithStdOutWriter(os.Stdout).WithAllowFailure(true)
		if !Run("Checking if the 'postgres_data' volume exists", localEnvCtx, false, "%s", "podman volume exists postgres_data") {
			Run("Creating the 'postgres_data' volume", localEnvCtx, false, "%s", "podman volume create postgres_data")
		}

		// let's use it in the compose file, so that the DB data is persisted across container restarts
		return `
volumes:
    postgres_data:
        external: true
`
	}
	return ""
}

// ----------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------

func hasPostgreSQLDb(dbConfigs map[string]map[string]interface{}) bool {
	for _, dbConfig := range dbConfigs {
		if dbType, ok := dbConfig["type"].(string); ok && dbType == "postgresql" {
			return true
		}
	}
	return false
}

func getDbServicesPart(dbConfigs map[string]map[string]interface{}) string {
	dbServices := ""

	for dbName, dbConfig := range dbConfigs {
		if dbType, ok := dbConfig["type"].(string); ok && dbType == "postgresql" {
			dbServices += postgresDbService(dbName, dbConfig)
		}
	}

	return dbServices
}

func getDependOnDbServicesPart(dbConfigs map[string]map[string]interface{}) string {
	dependOnDbServices := ""

	for dbName, dbConfig := range dbConfigs {
		if dbType, ok := dbConfig["type"].(string); ok && dbType == "postgresql" {
			dependOnDbServices += fmt.Sprintf("\n"+`            %s_db:
                condition: service_healthy
`, dbName)
		}
	}

	return dependOnDbServices
}

func postgresDbService(dbName string, dbConfig map[string]interface{}) string {
	return fmt.Sprintf(`
    # Defines the PostgreSQL database service
    %[1]s_db:
        image: %[2]s
        environment:
            POSTGRESQL_DATABASE: %[3]s
            POSTGRESQL_USER: %[4]s
            POSTGRESQL_PASSWORD: %[5]s
            POSTGRESQL_ADMIN_PASSWORD: %[6]s
        volumes:
            - postgres_data:/var/lib/pgsql/data
        healthcheck:
            test: ["CMD-SHELL", "pg_isready -U %[4]s -d %[3]s"]
            interval: 2s
            timeout: 5s
            retries: 10
            start_period: 5s
        ports:
            - "5432:5432" # giving access outside these services
`,
		dbName,                                 // 1
		Config().API.LocalDev.DbImages[dbName], // 2
		core.GetValueFromMap(dbConfig, "database"),      // 3
		core.GetValueFromMap(dbConfig, "admin", "user"), // 4
		core.GetValueFromMap(dbConfig, "admin", "pass"), // 5
		core.GetValueFromMap(dbConfig, "admin", "pass"), // 6
	)
}
