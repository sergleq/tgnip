package main

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Locale представляет локализацию для определенного языка
type Locale struct {
	WelcomeMessage     string
	ProcessingMessage  string
	InvalidURLMessage  string
	ErrorProcessingMsg string
	ErrorSendingMsg    string
	SuccessMessage     string
}

// locales содержит все поддерживаемые языки
var locales = map[string]Locale{
	"ru": {
		WelcomeMessage: `Привет! Я бот для преобразования ссылок в markdown файлы.

Отправьте мне ссылку на веб-страницу, и я создам для вас markdown файл с очищенным содержимым.

Поддерживаемые форматы:
- Статьи и новости
- Блог-посты
- Документация

Просто отправьте ссылку!`,
		ProcessingMessage:  "⏳ Обрабатываю ссылку...",
		InvalidURLMessage:  "Пожалуйста, отправьте валидную ссылку на веб-страницу.",
		ErrorProcessingMsg: "❌ Не удалось обработать ссылку. Проверьте, что ссылка корректна и доступна.",
		ErrorSendingMsg:    "❌ Не удалось отправить файл.",
		SuccessMessage:     "✅ Файл успешно создан!",
	},
	"en": {
		WelcomeMessage: `Hello! I'm a bot for converting links to markdown files.

Send me a link to a web page, and I'll create a markdown file with cleaned content for you.

Supported formats:
- Articles and news
- Blog posts
- Documentation

Just send a link!`,
		ProcessingMessage:  "⏳ Processing link...",
		InvalidURLMessage:  "Please send a valid link to a web page.",
		ErrorProcessingMsg: "❌ Failed to process the link. Check that the link is correct and accessible.",
		ErrorSendingMsg:    "❌ Failed to send the file.",
		SuccessMessage:     "✅ File successfully created!",
	},
	"es": {
		WelcomeMessage: `¡Hola! Soy un bot para convertir enlaces a archivos markdown.

Envíame un enlace a una página web, y crearé un archivo markdown con contenido limpio para ti.

Formatos soportados:
- Artículos y noticias
- Publicaciones de blog
- Documentación

¡Solo envía un enlace!`,
		ProcessingMessage:  "⏳ Procesando enlace...",
		InvalidURLMessage:  "Por favor, envía un enlace válido a una página web.",
		ErrorProcessingMsg: "❌ No se pudo procesar el enlace. Verifica que el enlace sea correcto y accesible.",
		ErrorSendingMsg:    "❌ No se pudo enviar el archivo.",
		SuccessMessage:     "✅ ¡Archivo creado exitosamente!",
	},
	"fr": {
		WelcomeMessage: `Bonjour ! Je suis un bot pour convertir les liens en fichiers markdown.

Envoyez-moi un lien vers une page web, et je créerai un fichier markdown avec du contenu nettoyé pour vous.

Formats pris en charge :
- Articles et actualités
- Articles de blog
- Documentation

Envoyez simplement un lien !`,
		ProcessingMessage:  "⏳ Traitement du lien...",
		InvalidURLMessage:  "Veuillez envoyer un lien valide vers une page web.",
		ErrorProcessingMsg: "❌ Impossible de traiter le lien. Vérifiez que le lien est correct et accessible.",
		ErrorSendingMsg:    "❌ Impossible d'envoyer le fichier.",
		SuccessMessage:     "✅ Fichier créé avec succès !",
	},
	"de": {
		WelcomeMessage: `Hallo! Ich bin ein Bot zum Konvertieren von Links in Markdown-Dateien.

Senden Sie mir einen Link zu einer Webseite, und ich erstelle eine Markdown-Datei mit bereinigtem Inhalt für Sie.

Unterstützte Formate:
- Artikel und Nachrichten
- Blog-Beiträge
- Dokumentation

Senden Sie einfach einen Link!`,
		ProcessingMessage:  "⏳ Verarbeite Link...",
		InvalidURLMessage:  "Bitte senden Sie einen gültigen Link zu einer Webseite.",
		ErrorProcessingMsg: "❌ Link konnte nicht verarbeitet werden. Überprüfen Sie, ob der Link korrekt und zugänglich ist.",
		ErrorSendingMsg:    "❌ Datei konnte nicht gesendet werden.",
		SuccessMessage:     "✅ Datei erfolgreich erstellt!",
	},
	"it": {
		WelcomeMessage: `Ciao! Sono un bot per convertire i link in file markdown.

Mandami un link a una pagina web, e creerò un file markdown con contenuto pulito per te.

Formati supportati:
- Articoli e notizie
- Post del blog
- Documentazione

Invia semplicemente un link!`,
		ProcessingMessage:  "⏳ Elaborazione del link...",
		InvalidURLMessage:  "Per favore, invia un link valido a una pagina web.",
		ErrorProcessingMsg: "❌ Impossibile elaborare il link. Verifica che il link sia corretto e accessibile.",
		ErrorSendingMsg:    "❌ Impossibile inviare il file.",
		SuccessMessage:     "✅ File creato con successo!",
	},
	"pt": {
		WelcomeMessage: `Olá! Sou um bot para converter links em arquivos markdown.

Envie-me um link para uma página web, e eu criarei um arquivo markdown com conteúdo limpo para você.

Formatos suportados:
- Artigos e notícias
- Posts de blog
- Documentação

Apenas envie um link!`,
		ProcessingMessage:  "⏳ Processando link...",
		InvalidURLMessage:  "Por favor, envie um link válido para uma página web.",
		ErrorProcessingMsg: "❌ Falha ao processar o link. Verifique se o link está correto e acessível.",
		ErrorSendingMsg:    "❌ Falha ao enviar o arquivo.",
		SuccessMessage:     "✅ Arquivo criado com sucesso!",
	},
	"zh": {
		WelcomeMessage: `你好！我是一个将链接转换为markdown文件的机器人。

给我发送一个网页链接，我将为您创建一个包含清理内容的markdown文件。

支持的格式：
- 文章和新闻
- 博客文章
- 文档

只需发送链接即可！`,
		ProcessingMessage:  "⏳ 正在处理链接...",
		InvalidURLMessage:  "请发送一个有效的网页链接。",
		ErrorProcessingMsg: "❌ 无法处理链接。请检查链接是否正确且可访问。",
		ErrorSendingMsg:    "❌ 无法发送文件。",
		SuccessMessage:     "✅ 文件创建成功！",
	},
	"ja": {
		WelcomeMessage: `こんにちは！リンクをmarkdownファイルに変換するボットです。

ウェブページのリンクを送ってください。クリーンなコンテンツでmarkdownファイルを作成します。

サポートされている形式：
- 記事とニュース
- ブログ投稿
- ドキュメント

リンクを送るだけです！`,
		ProcessingMessage:  "⏳ リンクを処理中...",
		InvalidURLMessage:  "有効なウェブページのリンクを送ってください。",
		ErrorProcessingMsg: "❌ リンクの処理に失敗しました。リンクが正しく、アクセス可能かどうか確認してください。",
		ErrorSendingMsg:    "❌ ファイルの送信に失敗しました。",
		SuccessMessage:     "✅ ファイルが正常に作成されました！",
	},
}

// getLocale определяет язык пользователя и возвращает соответствующую локализацию
func getLocale(message *tgbotapi.Message) Locale {
	// Проверяем язык пользователя
	if message.From != nil && message.From.LanguageCode != "" {
		lang := strings.ToLower(message.From.LanguageCode)

		// Проверяем точное совпадение
		if locale, exists := locales[lang]; exists {
			return locale
		}

		// Проверяем основную часть языка (например, "en" для "en-US")
		if len(lang) >= 2 {
			mainLang := lang[:2]
			if locale, exists := locales[mainLang]; exists {
				return locale
			}
		}
	}

	// Возвращаем русский как язык по умолчанию
	return locales["ru"]
}

// getSupportedLanguages возвращает список поддерживаемых языков
func getSupportedLanguages() []string {
	languages := make([]string, 0, len(locales))
	for lang := range locales {
		languages = append(languages, lang)
	}
	return languages
}
