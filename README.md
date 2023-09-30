# webCrawler
--
Tiny web crawler in Go with a bit of extra functions:

Features
- recursive crawling of web pages (thanks to colly lib)
- downloads web pages (same as the original domain only; html, js, txt only)
- when downloading a page, if a file with the same name already exists, the app will automatically attempt to resume the download (thanks to grub lib).
- supports resumable downloads for interrupted sessions, managed via cralwer_state.txt file
- saves the state of downloaded URLs to continue crawling from where you left off
- writes logs to the file as well as stdout

Limitations
- downloads are happening for web pages for the original domain only;
- only html, js, txt content is supported for downloading
- resumable downloads may not always work flawlessly

Usage
To use the web crawler, follow these steps:

Install Go:
- ensure that you have Go installed on your system.

Clone the Repository

Build the Application:
- cd webCrawler
- run "go build"

Run the Web Crawler:
./webCrawler -url <start-url> -depth <max-depth>

Example:
        ./webCrawler -url crawler-test.com -depth 3