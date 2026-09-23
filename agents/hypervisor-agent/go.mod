module github.com/bosscloud/bosscloud/agents/hypervisor-agent

go 1.23.0

require (
	github.com/bosscloud/bosscloud/libs/go/agentprotocol v0.0.0
	github.com/bosscloud/bosscloud/libs/go/config v0.0.0
	github.com/google/uuid v1.6.0
)

replace (
	github.com/bosscloud/bosscloud/libs/go/agentprotocol => ../../libs/go/agentprotocol
	github.com/bosscloud/bosscloud/libs/go/config => ../../libs/go/config
)
