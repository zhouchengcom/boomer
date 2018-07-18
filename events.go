package glocust

// Events is core event bus instance of boomer
// var Events = EventBus.New()

func RequestSuccess(requestType string, name string, responseTime int64, responseLength int64) {
	// println(len(requestSuccessChannel))
	requestSuccessChannel <- &requestSuccess{
		requestType:    requestType,
		name:           name,
		responseTime:   responseTime,
		responseLength: responseLength,
	}
}

func RequestFailure(requestType string, name string, responseTime int64, exception string) {
	requestFailureChannel <- &requestFailure{
		requestType:  requestType,
		name:         name,
		responseTime: responseTime,
		error:        exception,
	}
}
