package events

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/generator/program/common"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
)

func GenerateEvents(ctx *model.GenerateCtx, program *idl.Idl) *File {
	file := helper.NewGoFile(ctx)

	for _, evt := range program.Events {
		identType := ctx.GetIdentifier(evt.Name)
		if identType == nil {
			panic("event " + evt.Name + " not found in IDL types")
		}

		var discriminator *[8]byte
		if evt.Discriminator != nil {
			discriminator = (*[8]byte)(evt.Discriminator)
		}

		file.Add(
			common.GenerateTypeDefCode(
				ctx,
				evt.Name+"EventData",
				identType,
				discriminator,
				program,
			),
		)
	}

	file.Add(Empty().Var().Id("eventTypes").Op("=").Map(Index(Lit(8)).Byte()).Qual("reflect", "Type").Values(DictFunc(func(d Dict) {
		for _, evt := range program.Events {
			if identType := ctx.GetIdentifier(evt.Name); identType != nil {
				d[Id(evt.Name+"EventDataDiscriminator")] = Id("reflect.TypeOf(" + evt.Name + "EventData{})")
			}
		}
	})))

	file.Add(Empty().Var().Id("eventNames").Op("=").Map(Index(Lit(8)).Byte()).String().Values(DictFunc(func(d Dict) {
		for _, evt := range program.Events {
			if identType := ctx.GetIdentifier(evt.Name); identType != nil {
				d[Id(evt.Name+"EventDataDiscriminator")] = Lit(evt.Name)
			}
		}
	})))

	generateEventSnippet(file)

	return file
}
