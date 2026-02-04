package logging

import (
	"log/slog"
	"os"

	"github.com/makhkets/managment/internal/domain/logging"
	"github.com/makhkets/managment/pkg/lib/logger/handlers/slogpretty"
	"github.com/makhkets/managment/pkg/utils"
)

func SetupLogger() {
	//var log *slog.Logger
	setupPrettySlog()

	//case envProd:
	//	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	//		Level: slog.LevelInfo,
	//	}))
	//}

	//return log
}

func setupPrettySlog() {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	// Создание файла для логов
	file, err := os.OpenFile(
		utils.FindDirectoryName(
			"logger",
		)+
			"\\logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666,
	)
	if err != nil {
		panic(err)
	}
	//defer file.Close()
	customWriter := &model_logging.CustomFileWriter{File: file}
	handler := opts.NewPrettyHandler(os.Stdout, customWriter)

	slog.SetDefault(slog.New(handler))
}
