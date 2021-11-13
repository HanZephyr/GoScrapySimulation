package GoScrapySimulation

import (
	"log"
)

type Pipeline interface {
	StartPipeline(parameter interface{}) interface{}
	ProcessItemFunc(parameter interface{}, dataItem interface{}) interface{}
	ClosePipeline(parameter interface{}) interface{}
}


type DefaultPipeline struct{
	PipelineName string
}

func (defaultPipeline *DefaultPipeline) StartPipeline(parameter interface{}) interface{} { return nil }
func (defaultPipeline *DefaultPipeline) ProcessItemFunc(parameter interface{}, dataItem interface{}) interface{} {
	log.Printf("ProcessItem: %v", dataItem)
	return nil
}
func (defaultPipeline *DefaultPipeline) ClosePipeline(parameter interface{}) interface{} { return nil }
