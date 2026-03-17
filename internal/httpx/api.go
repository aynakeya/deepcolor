package httpx

type ParameterType any
type RequestFunc[P ParameterType] func(params P) (*Request, error)
type ParserResultType any
type ParserFunc[P ParserResultType] func(response *Response) (P, error)
type ResultType any
type ResultFunc[P ParserResultType, R ResultType] func(result P, container R) error
type NextRequestFunc[P ParameterType, T ParserResultType, R ResultType] func(params P, parsedResult T, result R) (nextParams P, ok bool)

type ApiInfo[P ParameterType, T ParserResultType, R ResultType] struct {
	Request RequestFunc[P]
	Parser  ParserFunc[T]
	Result  ResultFunc[T, R]
	// Next is optional, if Next exists, then it will be a sequence of requests
	Next NextRequestFunc[P, T, R]
}

func (api *ApiInfo[P, T, R]) Run(requester IRequester, para P, result R) error {
	var hasNext = true
	for hasNext {
		req, err := api.Request(para)
		if err != nil {
			return err
		}
		httpResp, err := requester.Do(req)
		if err != nil {
			return err
		}
		parsedResult, err := api.Parser(httpResp)
		if err != nil {
			return err
		}
		err = api.Result(parsedResult, result)
		if err != nil {
			return err
		}
		if api.Next == nil {
			hasNext = false
		} else {
			para, hasNext = api.Next(para, parsedResult, result)
		}
	}
	return nil
}

type ApiResultFunc[P ParameterType, R ResultType] func(P) (R, error)
type ApiRecverFunc[P ParameterType, R ResultType] func(P, R) error

func NewResultFunc[P ParameterType, T ParserResultType, R ResultType](
	requester IRequester, api *ApiInfo[P, T, *R]) ApiResultFunc[P, R] {
	apiFunc := NewRecverFunc(requester, api)
	return func(para P) (R, error) {
		var result R
		return result, apiFunc(para, &result)
	}
}

func NewRecverFunc[P ParameterType, T ParserResultType, R ResultType](
	requester IRequester, api *ApiInfo[P, T, R]) ApiRecverFunc[P, R] {
	if requester == nil {
		panic("can't make api, requester is nil")
	}
	if api.Request == nil ||
		api.Parser == nil ||
		api.Result == nil {
		panic("can't make api, ApiInfo missing attribute")
	}
	return func(para P, result R) error {
		return api.Run(requester, para, result)
	}
}
