package repositories

type ExampleRepository interface {
	GetExampleRecord() string
	GetSumQuery() int
}
