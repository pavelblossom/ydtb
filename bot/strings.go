package bot

const (
	greetingMessage = `
Send me a song's name or a link. More info /help
Note that I can't download too long songs (approx more than 10-15 min) dut to telegram restrictions to upload file'`

	commandsImplementErr = `
Command is not recognized. Please type / and select from existing`

	callbackImplementErr = `
An internal error occurred. Please check the duration of downloading (in most cases that's the problem) and
retry your request or if above not helps contact the author  if any issues or /help to get more info`

	downloadStarted = "Downloading..."

	messageImplementErr = `
Something went wrong and I couldn't download. Try next:
1. Check the URL that you've sent me (open it in browser)
2. Some services are not available. Check support status at /help
3. Internal error occurred. Contact with author to fix this error. Thanks!`

	helpMessage = `
*Search and download*
Write me a track name and author to find music. List of available downloads will appear.
In this case I download only from Youtube

*Download using link*
Send me a link to instant start downloading (youtube, soundcloud, vimeo, etc). Note: not all services are supported.
List of [available services](http://ytdl-org.github.io/youtube-dl/supportedsites.html)

- Backend is open-source (MIT license)
- Written on Golang
- Download using [youtube-dl](https://github.com/ytdl-org/youtube-dl)`
)
