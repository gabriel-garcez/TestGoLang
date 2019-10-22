package suite

import (
	"fmt"
	"log"
	"time"
)

func NewSystemLogs(verbose bool) (sysLog *SystemLogs) {
	sysLog = &SystemLogs{
		logHub:      make(chan *LogDetail),
		verboseMode: verbose,
	}
	go sysLog.run()
	return
}

func (sl *SystemLogs) run() {
	defer close(sl.logHub)
	for {
		select {
		case logMess := <-sl.logHub:
			go sl.registerLog(logMess)
		}
	}
}

func (sl *SystemLogs) registerLog(logDetail *LogDetail) {
	sl.LogHistory = append(sl.LogHistory, logDetail)
}

func (sl *SystemLogs) logMessageHub(logDetail *LogDetail) {
	sl.logHub <- logDetail
}

func (sl *SystemLogs) logMessageScreen(logDetail *LogDetail) {
	var message string

	if logDetail.EntityName != "" {
		message = fmt.Sprintf(" [Entity: %s]", logDetail.EntityName)
	}
	if logDetail.RoutineTitle != "" {
		message = fmt.Sprintf("%s [Routine: %s]", message, logDetail.RoutineTitle)
	}
	if logDetail.Message != "" {
		message = fmt.Sprintf("%s [Message: %s]", message, logDetail.Message)
	}
	if logDetail.Details != nil {
		message = fmt.Sprintf("%s [Details: %v]", message, logDetail.Details)
	}

	log.Println(generateMessage(message, logDetail.Level))
}

func generateMessage(msg, level string) string {
	switch level {
	case INFO:
		return fmt.Sprintf(InfoColor, "INFO", msg)
	case NOTICE:
		return fmt.Sprintf(NoticeColor, "NOTICE", msg)
	case ERROR:
		return fmt.Sprintf(ErrorColor, "ERROR", msg)
	case WARNING:
		return fmt.Sprintf(WarningColor, "WARNING", msg)
	case PANIC:
		log.Panic(fmt.Sprintf(ErrorColor, "PANIC", msg))
	case FATAL:
		log.Fatal(fmt.Sprintf(ErrorColor, "FATAL", msg))
	}
	return ""
}

// Log send log to save
func (sl *SystemLogs) Log(entityName, routineTitle, message, level string, details map[string]interface{}, forceLog bool) {
	newLog := &LogDetail{
		Date:         time.Now().UnixNano(),
		EntityName:   entityName,
		RoutineTitle: routineTitle,
		Message:      message,
		Level:        level,
		Details:      details,
	}
	go sl.logMessageHub(newLog)
	if sl.verboseMode || forceLog {
		sl.logMessageScreen(newLog)
	}
}
