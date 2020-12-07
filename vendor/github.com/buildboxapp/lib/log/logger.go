// обертка для логирования, которая дополняем аттрибутами логируемого процесса logrus
// дополняем значениями, идентифицирующими запущенный сервис UID,Name,Service

package log

import (
	"context"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"

	"fmt"
	"io"
	"os"
	"path/filepath"
)

var logrusB = logrus.New()
var sep = string(filepath.Separator)

type Log struct {

	// куда логируем? stdout/;*os.File на файл, в который будем писать логи
	Output io.Writer `json:"output"`
	//Debug:
	// сообщения отладки, профилирования.
	// В production системе обычно сообщения этого уровня включаются при первоначальном
	// запуске системы или для поиска узких мест (bottleneck-ов).

	//Info: - логировать процесс выполнения
	// обычные сообщения, информирующие о действиях системы.
	// Реагировать на такие сообщения вообще не надо, но они могут помочь, например,
	// при поиске багов, расследовании интересных ситуаций итд.

	//Warning: - логировать странные операции
	// записывая такое сообщение, система пытается привлечь внимание обслуживающего персонала.
	// Произошло что-то странное. Возможно, это новый тип ситуации, ещё не известный системе.
	// Следует разобраться в том, что произошло, что это означает, и отнести ситуацию либо к
	// инфо-сообщению, либо к ошибке. Соответственно, придётся доработать код обработки таких ситуаций.

	//Error: - логировать ошибки
	// ошибка в работе системы, требующая вмешательства. Что-то не сохранилось, что-то отвалилось.
	// Необходимо принимать меры довольно быстро! Ошибки этого уровня и выше требуют немедленной записи в лог,
	// чтобы ускорить реакцию на них. Нужно понимать, что ошибка пользователя – это не ошибка системы.
	// Если пользователь ввёл в поле -1, где это не предполагалось – не надо писать об этом в лог ошибок.

	//Panic: - логировать критические ошибки
	// это особый класс ошибок. Такие ошибки приводят к неработоспособности системы в целом, или
	// неработоспособности одной из подсистем. Чаще всего случаются фатальные ошибки из-за неверной конфигурации
	// или отказов оборудования. Требуют срочной, немедленной реакции. Возможно, следует предусмотреть уведомление о таких ошибках по SMS.
	// указываем уровни логирования Error/Warning/Debug/Info/Panic

	//Trace: - логировать обработки запросов

	// можно указывать через | разные уровени логирования, например Error|Warning
	// можно указать All - логирование всех уровней
	Levels string `json:"levels"`
	// uid процесса (сервиса), который логируется (случайная величина)
	UID string `json:"uid"`
	// имя процесса (сервиса), который логируется
	Name string `json:"name"`
	// название сервиса (app/gui...)
	Service string `json:"service"`
	// директория сохранения логов
	Dir string `json:"dir"`
	// uid-конфигурации с которой был запущен процесс
	Config string `json:"config"`
	// интервал между проверками актуального файла логирования (для текущего дня)
	IntervalReload time.Duration `json:"delay_reload"`
	// интервал проверками на наличие файлов на удаление
	IntervalClearFiles time.Duration `json:"interval_clear_files"`
	// период хранения файлов лет-месяцев-дней (например: 0-1-0 - хранить 1 месяц)
	PeriodSaveFiles string `json:"period_save_files"`
}

func (c *Log) Trace(args ...interface{}) {
	if strings.Contains(c.Levels, "Trace") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.TraceLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   c.Name,
			"uid":    c.UID,
			"srv":    c.Service,
			"config": c.Config,
		}).Trace(args...)
	}
}

func (c *Log) Debug(args ...interface{}) {
	if strings.Contains(c.Levels, "Debug") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		// Only log the warning severity or above.
		logrusB.SetLevel(logrus.DebugLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   c.Name,
			"uid":    c.UID,
			"srv":    c.Service,
			"config": c.Config,
		}).Debug(args...)
	}
}

func (c *Log) Info(args ...interface{}) {
	if strings.Contains(c.Levels, "Info") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		logrusB.SetLevel(logrus.InfoLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   c.Name,
			"uid":    c.UID,
			"srv":    c.Service,
			"config": c.Config,
		}).Info(args...)
	}
}

func (c *Log) Warning(args ...interface{}) {
	if strings.Contains(c.Levels, "Warning") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.WarnLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   c.Name,
			"uid":    c.UID,
			"srv":    c.Service,
			"config": c.Config,
		}).Warn(args...)
	}
}

func (c *Log) Error(err error, args ...interface{}) {
	if strings.Contains(c.Levels, "Error") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.ErrorLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   c.Name,
			"uid":    c.UID,
			"srv":    c.Service,
			"config": c.Config,
			"error":  fmt.Sprint(err),
		}).Error(args...)
	}
}

func (c *Log) Panic(err error, args ...interface{}) {
	if strings.Contains(c.Levels, "Panic") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.PanicLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   c.Name,
			"uid":    c.UID,
			"srv":    c.Service,
			"config": c.Config,
			"error":  fmt.Sprint(err),
		}).Panic(args...)
	}
}

// внутренняя ф-ция логирования и прекращения работы программы
func (c *Log) Exit(err error, args ...interface{}) {
	if strings.Contains(c.Levels, "Fatal") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})
		logrusB.SetLevel(logrus.FatalLevel)

		logrusB.WithFields(logrus.Fields{
			"name":   c.Name,
			"uid":    c.UID,
			"srv":    c.Service,
			"config": c.Config,
			"error":  fmt.Sprint(err),
		}).Fatal(args...)
	}
}

// Переинициализация файла логирования
func (c *Log) RotateInit(ctx context.Context) {

	// попытка обновить файл (раз в 10 минут)
	go func() {
		ticker := time.NewTicker(c.IntervalReload)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				b := New(c.Dir, c.Levels, c.UID, c.Name, c.Service, c.Config, c.IntervalReload, c.IntervalClearFiles, c.PeriodSaveFiles)
				c.Output = b.Output
				ticker = time.NewTicker(c.IntervalReload)
			}
		}
	}()

	// попытка очистки старых файлов (каждые пол часа)
	go func() {
		ticker := time.NewTicker(c.IntervalClearFiles)
		defer ticker.Stop()

		// получаем период, через который мы будем удалять файлы
		period := c.PeriodSaveFiles
		if period == "" {
			c.Error(fmt.Errorf("%s", "Fail perion save log files. (expected format: year-month-day; eg: 0-1-0)"))
			return
		}
		slPeriod := strings.Split(period, "-")
		if len(slPeriod) < 3 {
			c.Error(fmt.Errorf("%s", "Fail perion save log files. (expected format: year-month-day; eg: 0-1-0)"))
			return
		}

		// получаем числовые значения года месяца и дня для расчета даты удаления файлов
		year, err := strconv.Atoi(slPeriod[0])
		if err != nil {
			c.Error(err, "Fail converted Year from period saved log files. (expected format: year-month-day; eg: 0-1-0)")
		}
		month, err := strconv.Atoi(slPeriod[1])
		if err != nil {
			c.Error(err, "Fail converted Month from period saved log files. (expected format: year-month-day; eg: 0-1-0)")
		}
		day, err := strconv.Atoi(slPeriod[2])
		if err != nil {
			c.Error(err, "Fail converted Day from period saved log files. (expected format: year-month-day; eg: 0-1-0)")
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				oneMonthAgo := time.Now().AddDate(-year, -month, -day) // minus 1 месяц
				fileMonthAgoDate := oneMonthAgo.Format("2006.01.02")

				// пробегаем директорию и читаем все файлы, если имя меньше текущее время - месяц = удаляем
				directory, _ := os.Open(c.Dir)
				objects, err := directory.Readdir(-1)
				if err != nil {
					c.Error(err, "Error read directory: ", directory)
					return
				}

				for _, obj := range objects {
					filename := obj.Name()
					filenameMonthAgoDate := fileMonthAgoDate + "_" + c.Service

					if filenameMonthAgoDate > filename {
						pathFile := c.Dir + sep + filename
						err = os.Remove(pathFile)
						if err != nil {
							c.Error(err, "Error deleted file: ", pathFile)
							return
						}
					}
				}
				ticker = time.NewTicker(c.IntervalClearFiles)
			}
		}
	}()
}

func New(logsDir, level, uid, name, srv, config string, intervalReload, intervalClearFiles time.Duration, periodSaveFiles string) *Log {
	var output io.Writer
	var err error
	var mode os.FileMode

	datefile := time.Now().Format("2006.01.02")
	logName := datefile + "_" + srv + ".log"

	// создаем/открываем файл логирования и назначаем его логеру
	mode = 0711
	err = os.MkdirAll(logsDir, mode)
	if err != nil {
		logrus.Error(err, "Error creating directory")
		return nil
	}

	output, err = os.OpenFile(logsDir+"/"+logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logrus.Panic(err, "error opening file")
		return nil
	}

	return &Log{
		Output:             output,
		Levels:             level,
		UID:                uid,
		Name:               name,
		Service:            srv,
		Dir:                logsDir,
		Config:             config,
		IntervalReload:     intervalReload,
		IntervalClearFiles: intervalClearFiles,
		PeriodSaveFiles:    periodSaveFiles,
	}
}
