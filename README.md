# webCrawler

Tiny web crawler in Go with a bit of extra functions:

# Features
- recursive crawling of web pages (thanks to colly lib)
- downloads web pages (same as the original domain only; html, js, txt only)
- supports resumable downloads for interrupted sessions (thanks to grub lib)
- saves the state of downloaded URLs (managed via cralwer_state.txt file) to continue crawling from where you left off
- writes logs to the file as well as stdout

# Limitations
- downloads are happening for web pages for the original domain only
- only html, js, txt content is supported for downloading
- when downloading a page, if a file with the same name already exists, the app will automatically attempt to resume the download. however this may not always work flawlessly

# Usage
To use the web crawler, follow these steps:

1. Install Go:\
   ensure that you have Go installed on your system.

2. Clone the repo:\
   git@github.com:svasina/tinyWebCrawler.git

3. Build the app:\
   cd webCrawler\
   run "go build"

4. Run the Tiny Web Crawler:\
        ./webCrawler -url \<start-url\> -depth \<max-depth\>\
        (default depth is 3)\
\
Example:\
        ./webCrawler -url crawler-test.com -depth 3
