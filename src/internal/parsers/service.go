package parsers

type service struct{}

func NewParserService() ParserService {
	return &service{}
}
