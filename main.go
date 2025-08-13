package main

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// MessageTask представляет задачу обработки сообщения
type MessageTask struct {
	Update tgbotapi.Update
	Bot    *tgbotapi.BotAPI
}

// Metrics представляет метрики производительности
type Metrics struct {
	TotalRequests     int64
	ActiveRequests    int64
	CompletedRequests int64
	FailedRequests    int64
	StartTime         time.Time
}

// WorkerPool представляет пул воркеров для обработки сообщений
type WorkerPool struct {
	workers     int
	taskChan    chan MessageTask
	wg          sync.WaitGroup
	bot         *tgbotapi.BotAPI
	metrics     *Metrics
	rateLimiter *time.Ticker // Rate limiter для Telegram API
}

// NewWorkerPool создает новый пул воркеров
func NewWorkerPool(config *Config, bot *tgbotapi.BotAPI) *WorkerPool {
	rateLimitInterval := time.Duration(config.GetRateLimitInterval()) * time.Millisecond

	return &WorkerPool{
		workers:     config.WorkerCount,
		taskChan:    make(chan MessageTask, config.QueueBufferSize),
		bot:         bot,
		metrics:     &Metrics{StartTime: time.Now()},
		rateLimiter: time.NewTicker(rateLimitInterval),
	}
}

// Start запускает пул воркеров
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	logger.Infof("Запущен пул из %d воркеров", wp.workers)

	// Запускаем горутину для логирования метрик
	go wp.logMetrics()
}

// Stop останавливает пул воркеров
func (wp *WorkerPool) Stop() {
	close(wp.taskChan)
	wp.wg.Wait()
	wp.rateLimiter.Stop()
	logger.Info("Пул воркеров остановлен")
}

// Submit добавляет задачу в очередь
func (wp *WorkerPool) Submit(task MessageTask) {
	atomic.AddInt64(&wp.metrics.TotalRequests, 1)

	select {
	case wp.taskChan <- task:
		// Задача добавлена в очередь
		logger.Debugf("Задача добавлена в очередь, всего в очереди: %d", len(wp.taskChan))
	default:
		// Очередь переполнена, отправляем сообщение о перегрузке
		logger.Warn("Очередь задач переполнена, отправляем сообщение о перегрузке")
		atomic.AddInt64(&wp.metrics.FailedRequests, 1)
		if task.Update.Message != nil {
			locale := getLocale(task.Update.Message)
			msg := tgbotapi.NewMessage(task.Update.Message.Chat.ID, locale.ServerOverloadMessage)
			task.Bot.Send(msg)
		}
	}
}

// worker обрабатывает задачи из очереди
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	logger.Infof("Воркер %d запущен", id)

	for task := range wp.taskChan {
		atomic.AddInt64(&wp.metrics.ActiveRequests, 1)
		logger.Infof("Воркер %d обрабатывает сообщение от пользователя %d", id, task.Update.Message.Chat.ID)

		// Rate limiting для Telegram API
		<-wp.rateLimiter.C

		// Обработка команды /start
		if task.Update.Message.IsCommand() && task.Update.Message.Command() == "start" {
			handleStartCommand(task.Bot, task.Update.Message)
			atomic.AddInt64(&wp.metrics.CompletedRequests, 1)
			atomic.AddInt64(&wp.metrics.ActiveRequests, -1)
			continue
		}

		// Обработка ссылок
		if task.Update.Message.Text != "" {
			handleURLMessage(task.Bot, task.Update.Message)
			atomic.AddInt64(&wp.metrics.CompletedRequests, 1)
		}

		atomic.AddInt64(&wp.metrics.ActiveRequests, -1)
	}

	logger.Infof("Воркер %d завершен", id)
}

// logMetrics периодически логирует метрики производительности
func (wp *WorkerPool) logMetrics() {
	config := LoadConfig()
	ticker := time.NewTicker(time.Duration(config.MetricsInterval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		uptime := time.Since(wp.metrics.StartTime)
		total := atomic.LoadInt64(&wp.metrics.TotalRequests)
		active := atomic.LoadInt64(&wp.metrics.ActiveRequests)
		completed := atomic.LoadInt64(&wp.metrics.CompletedRequests)
		failed := atomic.LoadInt64(&wp.metrics.FailedRequests)

		logger.Infof("Метрики: Uptime=%v, Total=%d, Active=%d, Completed=%d, Failed=%d, Queue=%d",
			uptime, total, active, completed, failed, len(wp.taskChan))
	}
}

func init() {
	// Инициализация логгера
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		logger.Warn("Файл .env не найден, используем переменные окружения системы")
	}
}

func main() {
	// Загрузка конфигурации
	config := LoadConfig()

	// Настройка уровня логирования
	switch config.LogLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Получение токена бота
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		logger.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Создание бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Fatal(err)
	}

	bot.Debug = false
	logger.Infof("Бот %s запущен", bot.Self.UserName)
	logger.Infof("Конфигурация: Workers=%d, QueueSize=%d, RateLimit=%d/sec",
		config.WorkerCount, config.QueueBufferSize, config.RateLimitPerSec)

	// Создание пула воркеров
	workerPool := NewWorkerPool(config, bot)
	workerPool.Start()
	defer workerPool.Stop()

	// Настройка обновлений
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// Обработка сообщений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Создаем задачу и отправляем в пул воркеров
		task := MessageTask{
			Update: update,
			Bot:    bot,
		}
		workerPool.Submit(task)
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	locale := getLocale(message)
	msg := tgbotapi.NewMessage(message.Chat.ID, locale.WelcomeMessage)
	bot.Send(msg)
}

func handleURLMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	locale := getLocale(message)
	url := message.Text

	// Проверка, что это действительно ссылка
	if !isValidURL(url) {
		msg := tgbotapi.NewMessage(message.Chat.ID, locale.InvalidURLMessage)
		bot.Send(msg)
		return
	}

	// Отправляем сообщение о начале обработки
	processingMsg := tgbotapi.NewMessage(message.Chat.ID, locale.ProcessingMessage)
	sentMsg, err := bot.Send(processingMsg)
	if err != nil {
		logger.Errorf("Ошибка при отправке сообщения о обработке: %v", err)
	}

	// Извлекаем контент
	content, err := extractContent(url)
	if err != nil {
		logger.Errorf("Ошибка при извлечении контента: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, locale.ErrorProcessingMsg)
		bot.Send(errorMsg)
		return
	}

	// Конвертируем в markdown
	markdown := convertToMarkdown(content, url, locale)

	// Создаем файл
	filename := generateFilename(url, content.Title)
	file := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FileBytes{
		Name:  filename,
		Bytes: []byte(markdown),
	})

	// Отправляем файл
	if _, err := bot.Send(file); err != nil {
		logger.Errorf("Ошибка при отправке файла: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, locale.ErrorSendingMsg)
		bot.Send(errorMsg)
		return
	}

	// Удаляем сообщение о обработке
	if sentMsg.MessageID != 0 {
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, sentMsg.MessageID)
		bot.Send(deleteMsg)
	}

	logger.Infof("Файл успешно отправлен пользователю %d", message.Chat.ID)
}
