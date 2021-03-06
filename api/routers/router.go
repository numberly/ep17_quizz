package routers

import (
	"ep17_quizz/api/handlers"

	"github.com/julienschmidt/httprouter"
)

// Basic regroup handlers : Panic, Mongo
func Basic(h httprouter.Handle) httprouter.Handle {
	return handlers.Logging(
		handlers.Panic(
			handlers.Rethink(h),
		),
	)
}
