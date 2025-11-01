package testdata

type MockService struct {
	CallCount int
}

func (m *MockService) DoSomething() {
	m.CallCount++
}
