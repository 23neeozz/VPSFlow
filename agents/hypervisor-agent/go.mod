module github.com/vpsflow/vpsflow/agents/hypervisor-agent

go 1.23.0

require (
	github.com/vpsflow/vpsflow/libs/go/agentprotocol v0.0.0
	github.com/vpsflow/vpsflow/libs/go/config v0.0.0
	github.com/google/uuid v1.6.0
)

replace (
	github.com/vpsflow/vpsflow/libs/go/agentprotocol => ../../libs/go/agentprotocol
	github.com/vpsflow/vpsflow/libs/go/config => ../../libs/go/config
)
