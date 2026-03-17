package httpx

func FetchParsedResult[P ParserResultType](requester IRequester, request *Request, parserFunc ParserFunc[P]) (P, error) {
	httpResp, err := requester.Do(request)
	if err != nil {
		return *new(P), err
	}
	return parserFunc(httpResp)
}
