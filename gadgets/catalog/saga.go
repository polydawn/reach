package catalog

type SagaName struct{ s string }

var SagaNameZero = SagaName{}

func ParseSagaName(s string) (*SagaName, error) {
	// TODO check boundedness to [a-z0-9-], etc
	return &SagaName{s}, nil
}

func (sn SagaName) String() string {
	return sn.s
}
