package exportentities

import (
	"testing"

	"github.com/ovh/cds/engine/api/test"
	"github.com/ovh/cds/sdk"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestApplication_Application(t *testing.T) {

	var yml = `name: cds-github
repo_manager: github
repo_name: ovh/cds
permissions:
  CDS_TEAM: 7
variables:
  gitBranch:
    type: string
    value: '{{.git.branch}}'
  gitUrl:
    type: string
    value: https://github.com/ovh/cds
  marathon.config:
    type: text
    value: |-
      {
          "id": "/{{.cds.env.name}}/api",
          "mem": 2048,
          "cpus": 1,
          "instances": {{.cds.env.instances}},
          "upgradeStrategy": {
              "minimumHealthCapacity": 1,
              "maximumOverCapacity": 1
          },
          "container": {
              "type": "DOCKER",
              "docker": {
                  "network": "BRIDGE",
                  "portMappings": [
                      {
                          "containerPort": 8081,
                          "hostPort": 0,
                          "protocol": "tcp"
                      }
                  ],
                  "image": "registry.ovh.net/cds/engine/api:{{.cds.version}}",
                  "forcePullImage": false
              }
          },
          "env": {
              "CDS_API_URL": "{{.cds.env.apiBaseURL}}",
              "CDS_ARTIFACT_MODE": "{{.cds.env.artifact.mode}}",
              "CDS_ARTIFACT_BASEDIR": "{{.cds.env.artifact.basedir}}",
              "CDS_ARTIFACT_ADDRESS": "{{.cds.env.openstackAddress}}",
              "CDS_ARTIFACT_USER": "{{.cds.env.openstackUser}}",
              "CDS_ARTIFACT_PASSWORD": "{{.cds.env.openstackPassword}}",
              "CDS_ARTIFACT_TENANT": "{{.cds.env.openstackTenant}}",
              "CDS_ARTIFACT_REGION": "{{.cds.env.openstackRegion}}",
              "CDS_BASE_URL": "{{.cds.env.uiBaseURL}}",
              "CDS_DB_SECRET": "{{.cds.env.dbSecret}}",
              "CDS_DB_LOGGING": "0",
              "CDS_DB_HOST": "{{.cds.env.dbHost}}",
              "CDS_DB_NAME": "{{.cds.env.dbName}}",
              "CDS_DB_PASSWORD": "{{.cds.env.dbPassword}}",
              "CDS_DB_PORT": "{{.cds.env.api.db_port}}",
              "CDS_DB_USER": "{{.cds.env.dbUser}}",
              "CDS_DB_MAXCONN": "{{.cds.env.maxDbConns}}",
              "CDS_DB_TIMEOUT": "{{.cds.env.statementTimeout}}",
              "CDS_ARCHIVED_BUILD_HOURS": "240",
              "CDS_LISTEN_PORT": "8081",
              "CDS_NO_SMTP": "{{.cds.env.api.no_smtp}}",
              "CDS_SMTP_FROM": "noreply.cds@interne.ovh.net",
              "CDS_SMTP_HOST": "{{.cds.env.api.smtp}}",
              "CDS_SMTP_PORT": "25",
              "CDS_AUTH_LOCAL_MODE": "{{.cds.env.auth_mode}}",
              "CDS_LOG_LEVEL": "{{.cds.env.logLevel}}",
              "CDS_CACHE": "{{.cds.env.api.cache}}",
              "CDS_REDIS_HOST": "{{.cds.env.redisHost}}",
              "CDS_REDIS_PASSWORD": "{{.cds.env.redisPassword}}",
              "CDS_NO_REPO_POLLING":"{{.cds.env.api.polling.disable}}",
              "CDS_NO_REPO_CACHE_LOADER":"{{.cds.env.api.repo.disable}}",
              "CDS_NO_SCHEDULER":"{{.cds.env.api.no_scheduler}}",
              "CDS_NO_GITHUB_STATUS_URL":"true",
              "CDS_NO_GITHUB_STATUS":"{{.cds.env.api.github.status_disabled}}",
              "CDS_NO_STASH_STATUS":"{{.cds.env.api.stash.status_disabled}}",
              "CDS_EVENT_KAFKA_BROKER_ADDRESSES": "{{.cds.env.kafka.brokers}}",
              "CDS_EVENT_KAFKA_ENABLED":"{{.cds.env.kafka.enabled}}",
              "CDS_EVENT_KAFKA_TOPIC": "{{.cds.env.kafka.topic.events}}",
              "CDS_EVENT_KAFKA_USER": "{{.cds.env.kafka.writer.user}}",
              "CDS_EVENT_KAFKA_PASSWORD": "{{.cds.env.kafka.writer.password}}",
              "CDS_DEFAULT_GROUP": "{{.cds.env.api.defaultGroup}}",
              "VAULT_APPLICATION_KEY": "{{.cds.env.vaultApplicationKey}}",
              "VAULT_ADDR":"{{.cds.env.vaultAddr}}",
              "VERSION": "master{{.cds.version}}"
          },
          "labels": {
              "VAULT_ENABLE": "1",
              "USER_THOT_TOKEN": "{{.cds.env.thotToken}}",
              "HAPROXY_0_MODE": "http",
              "LB_0_MODE": "http",
              "HAPROXY_0_VHOST": "{{.cds.env.apiDomain}}.{{.cds.env.vHost}}",
              "LB_0_VHOST": "{{.cds.env.apiDomain}}.{{.cds.env.vHost}}"
          },
          "healthChecks": [
              {
                  "path": "/mon/status",
                  "protocol": "HTTP",
                  "portIndex": 0,
                  "gracePeriodSeconds": 15,
                  "intervalSeconds": 60,
                  "timeoutSeconds": 10,
                  "maxConsecutiveFailures": 2,
                  "ignoreHttp1xx": false
              }
          ],
          "constraints": [["hostname", "GROUP_BY"]]
      }
  marathon.file:
    type: string
    value: marathon-api.json
  repo:
    type: string
    value: https://github.com/ovh/cds.git
  uiImageName:
    type: string
    value: cds/ui-ng2
pipelines:
  build-api-worker-hatchery-cli-github:
    triggers:
      deploy-marathon-app:
      - environment: p191-prod
        manual: true
        conditions:
        - variable: git.branch
          expected: master
      - environment: iad1-prod
        manual: true
        conditions:
        - variable: git.branch
          expected: statusMail
      - environment: p191-preprod
        manual: false
        conditions:
        - variable: git.branch
          expected: refactorEstimate
    options:
    - notifications:
        jabber:
          body: ""
          on_success: always
          recipients: cdev@conference.jabber.ovh.net
          subject: '{{.cds.application}} Build API[{{.cds.version}}] {{.cds.status}}
            on Branch : {{.git.branch}}'
  build-github:
    triggers:
      build-api-worker-hatchery-cli-github:
      - manual: false
      build-ui-ng2:
      - manual: false
        conditions:
        - variable: git.branch
          expected: ui.*
      build-ui-ng2-cache:
      - manual: false
        conditions:
        - variable: git.branch
          expected: ui.*|refactoNgModules
    options:
    - polling: true
  build-ui-ng2:
    triggers:
      deploy-ui2:
      - environment: p191-preprod
        manual: false
        conditions:
        - variable: git.branch
          expected: ui-master
    options:
    - notifications:
        jabber:
          body: ""
          on_success: always
          recipients: cdev@conference.jabber.ovh.net
          subject: '{{.cds.application}} Build UI[{{.cds.version}}] {{.cds.status}}
            on Branch : {{.git.branch}}'
  cds-integration:
    triggers:
      deploy-marathon-app:
      - from_environment: p191-preprod
        environment: p191-prod
        manual: true
        conditions:
        - variable: git.branch
          expected: master
    options:
    - environment: p191-preprod
      notifications:
        jabber:
          body: ""
          on_start: true
          on_success: always
          recipients: cdev@conference.jabber.ovh.net
          send_to_author: false
          subject: '{{.cds.application}} Integration Tests [{{.cds.version}}] {{.cds.status}}
            on Branch : {{.git.branch}}'
  deploy-marathon-app:
    triggers:
      deploy-marathon-app:
      - application_name: cds-hatchery-swarm
        from_environment: p191-preprod
        environment: p191-preprod
        manual: false
        conditions:
        - variable: git.branch
          expected: master|refactorEstimate
      - application_name: cds-hatchery-marathon
        from_environment: p191-prod
        environment: p191-prod
        manual: true
        conditions:
        - variable: git.branch
          expected: master
      - application_name: cds-hatchery-openstack
        from_environment: p191-prod
        environment: p191-prod
        manual: true
        conditions:
        - variable: git.branch
          expected: master
      - application_name: cds-hatchery-swarm
        from_environment: p191-prod
        environment: p191-prod
        manual: true
        conditions:
        - variable: git.branch
          expected: master
      - application_name: cds-hatchery-marathon
        from_environment: p191-preprod
        environment: p191-preprod
        manual: false
        conditions:
        - variable: git.branch
          expected: master|refactorEstimate
      - application_name: cds-hatchery-marathon
        from_environment: iad1-prod
        environment: iad1-prod
        manual: false
        conditions:
        - variable: git.branch
          expected: master
    options:
    - environment: p191-prod
      notifications:
        jabber:
          body: ""
          on_success: always
          recipients: cdev@conference.jabber.ovh.net
          send_to_author: false
          subject: '{{.cds.application}} Deployment[{{.cds.version}}] {{.cds.status}}
            on {{.cds.environment}} (branch: {{.git.branch}})'
    - environment: iad1-prod
      notifications:
        jabber:
          body: |-
            Status : {{.cds.status}}
            Branch : {{.git.branch}}
            Details : {{.cds.buildURL}}
          on_success: always
          recipients: cdev@conference.jabber.ovh.net
          send_to_groups: true
          subject: '{{.cds.project}}/{{.cds.application}} {{.cds.pipeline}} {{.cds.environment}}'
  deploy-redis-ha: {}`

	a := &Application{}
	test.NoError(t, yaml.Unmarshal([]byte(yml), a))

	app, err := a.Application()
	test.NoError(t, err)

	assert.NotNil(t, app)

	assert.Equal(t, a.Name, app.Name)
	assert.Equal(t, a.RepositoryManager, app.RepositoriesManager.Name)
	assert.Equal(t, a.RepositoryName, app.RepositoryFullname)

	assert.Equal(t, len(a.Permissions), len(app.ApplicationGroups))
	assert.Equal(t, "CDS_TEAM", app.ApplicationGroups[0].Group.Name)
	assert.Equal(t, 7, app.ApplicationGroups[0].Permission)

	assert.Equal(t, 6, len(app.Variable))
	var checkVariable = func(v string) {
		appvar := sdk.VariableFind(app.Variable, v)
		assert.NotNil(t, appvar, "variable %s not found", v)
		if appvar != nil {
			assert.Equal(t, a.Variables[v].Value, appvar.Value)
		}
	}

	checkVariable("gitBranch")
	checkVariable("gitUrl")
	checkVariable("marathon.config")
	checkVariable("marathon.file")
	checkVariable("repo")
	checkVariable("uiImageName")

	var checkPipeline = func(p string) {
		var found bool
		for _, pip := range app.Pipelines {
			if pip.Pipeline.Name == p {
				found = true
				break
			}
		}
		assert.True(t, found, "pipeline %s not found", p)
	}

	checkPipeline("build-api-worker-hatchery-cli-github")
	checkPipeline("build-github")
	checkPipeline("build-ui-ng2")
	checkPipeline("build-ui-ng2-cache")
	checkPipeline("cds-integration")
	checkPipeline("deploy-marathon-app")

}
