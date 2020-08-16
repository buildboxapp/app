// обертка для логирования, которая дополняем аттрибутами логируемого процесса logrus
// дополняем значениями, идентифицирующими запущенный сервис UID,Name,Service

package logger

import (
	"github.com/sirupsen/logrus"
	"io"
	"fmt"
)

var logrusB = logrus.New()

type Log struct {

	// куда логируем? stdout/;*os.File на файл, в который будем писать логи
	Output   		io.Writer `json:"output"`
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

	//Fatal: - логировать критические ошибки
	// это особый класс ошибок. Такие ошибки приводят к неработоспособности системы в целом, или
	// неработоспособности одной из подсистем. Чаще всего случаются фатальные ошибки из-за неверной конфигурации
	// или отказов оборудования. Требуют срочной, немедленной реакции. Возможно, следует предусмотреть уведомление о таких ошибках по SMS.
	// указываем уровни логирования Error/Warning/Debug/Info/Fatal

	//Trace: - логировать обработки запросов


	// можно указывать через | разные уровени логирования, например Error|Warning
	// можно указать All - логирование всех уровней
	Levels 		string `json:"levels"`
	// uid процесса (сервиса), который логируется
	UID 	string `json:"uid"`
	// имя процесса (сервиса), который логируется
	Name string `json:"name"`
	// название сервиса (app/gui...)
	Service string `json:"service"`
}

func (c *Log) Init(output io.Writer, level, uid, name, srv string) {
	c.Output = output
	c.Levels = level
	c.UID = uid
	c.Name = name
	c.Service = srv
}

func (c *Log) Trace(args ...interface{}) {

	logrusB.SetOutput(c.Output)
	logrusB.SetFormatter(&logrus.JSONFormatter{})

	logrusB.WithFields(logrus.Fields{
		"name": 	c.Name,
		"uid":   	c.UID,
		"srv":   	c.Service,
	}).Trace(args...)

}


func (c *Log) Debug(args ...interface{}) {

	logrusB.SetOutput(c.Output)
	logrusB.SetFormatter(&logrus.JSONFormatter{})

	// Only log the warning severity or above.
	//logrusB.SetLevel(logrus.InfoLevel)

	logrusB.WithFields(logrus.Fields{
		"name": 	c.Name,
		"uid":   	c.UID,
		"srv":   	c.Service,
	}).Debug(args...)
}

func (c *Log) Info(args ...interface{}) {

	logrusB.SetOutput(c.Output)
	logrusB.SetFormatter(&logrus.JSONFormatter{})

	logrusB.WithFields(logrus.Fields{
		"name": 	c.Name,
		"uid":   	c.UID,
		"srv":   	c.Service,
	}).Info(args...)

}

func (c *Log) Warning(args ...interface{}) {

	logrusB.SetOutput(c.Output)
	logrusB.SetFormatter(&logrus.JSONFormatter{})

	logrusB.WithFields(logrus.Fields{
		"name": 	c.Name,
		"uid":   	c.UID,
		"srv":   	c.Service,
	}).Warn(args...)

}

func (c *Log) Error(err error, args ...interface{}) {

	logrusB.SetOutput(c.Output)
	logrusB.SetFormatter(&logrus.JSONFormatter{})

	logrusB.WithFields(logrus.Fields{
		"name": 	c.Name,
		"uid":   	c.UID,
		"srv":   	c.Service,
		"error":	fmt.Sprint(err),
	}).Error(args...)

}

func (c *Log) Fatal(err error, args ...interface{}) {

	logrusB.SetOutput(c.Output)
	logrusB.SetFormatter(&logrus.JSONFormatter{})

	logrusB.WithFields(logrus.Fields{
		"name": 	c.Name,
		"uid":   	c.UID,
		"srv":   	c.Service,
		"error":	fmt.Sprint(err),
	}).Fatal(args...)

}