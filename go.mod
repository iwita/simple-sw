module github.com/iwita/simple-sw

go 1.16

//replace github.com/serverlessworkflow/sdk-go v1.0.0 => /home/achilleas/go/src/github.com/serverlessworkflow/sdk-go

//replace github.com/iwita/simple-sw/pkg/runtime v0.0.0-20210728153612-0d1c8539cde9 => /home/achilleas/go/src/github.com/iwita/simple-sw/pkg/runtime

replace github.com/serverlessworkflow/sdk-go v1.0.0 => /home/dgiagos/goprojects/sdk-go

require (
	github.com/apache/openwhisk-client-go v0.0.0-20210313152306-ea317ea2794c
	github.com/itchyny/gojq v0.12.4
	github.com/serverlessworkflow/sdk-go v1.0.0
)
