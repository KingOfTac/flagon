package models

// Interfaces

type Definition struct{}

type Configuration struct{}

type ConfigurationFile struct{}

type Command struct{}

type Argument struct{}

type GlobalFlag struct{}

type PositionalFlag struct{}

type LocalFlag struct{}

type Operation struct{}

type Example struct{}

// Handler Types

type BeforeHandlerProps struct{}

type HandlerProps struct{}

type AfterHandlerProps struct{}

type SuccessHandlerProps struct{}

type FailureHandlerProps struct{}

type OperationResult struct{}

type BeforeHandler = func(props BeforeHandlerProps) error

type Handler = func(props HandlerProps) error

type AfterHandler = func(props AfterHandlerProps) error

type SuccessHandler = func(props SuccessHandlerProps) error

type FailureHandler = func(props FailureHandlerProps) error

type OperationHandler = func(props HandlerProps) (OperationResult, error)

// Configuration

type GetConfigurationFile = func(id string) (ConfigFile[any], error)

type GetContent[T any] = struct {
	data     T
	error    error
	hasError bool
}

type SetContent = struct {
	error    error
	hasError bool
}

type ConfigFile[T any] = struct {
	getContent func() (GetContent[T], error)
	setContent func(data T) SetContent
}
