import asyncio
from aiohttp import ClientSession
from urllib.parse import urlparse
from crawl import normalize_url, extract_page_data, get_urls_from_html

async def crawl_site_async(base_url, max_concurrency, max_pages):
    crawler = AsyncCrawler(base_url, max_concurrency, max_pages)
    async with crawler:
        return await crawler.crawl()

class AsyncCrawler:
    def __init__(self, base_url, max_concurrency, max_pages):
        self.base_url = base_url
        self.base_domain = urlparse(base_url).netloc
        self.page_data = {}
        self.lock = asyncio.Lock()
        self.max_concurrency = max_concurrency
        self.max_pages = max_pages
        self.should_stop = False
        self.all_tasks = set()
        self.semaphore = asyncio.Semaphore(max_concurrency)
        self.session = None
    
    async def __aenter__(self):
        self.session = ClientSession()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.session.close()

    async def add_page_visit(self, normalized_url):
        async with self.lock:
            if self.should_stop:
                return False
            if normalized_url in self.page_data:
                return False
            if len(self.page_data) >= self.max_pages:
                self.should_stop = True
                print("Reached maximum number of pages to crawl")
                
                for task in self.all_tasks:
                    if not task.done():
                        task.cancel()
               
                return False
            self.page_data[normalized_url] = None
            return True
    
    async def get_html(self, url):
        try:  
            resp = await self.session.get(url, headers={"User-Agent": "crawler"})
        except Exception as e:
            raise Exception(f"network error while fething {url}: {e}")
        async with resp:
            if resp.status > 399:
                raise Exception(f"invalid request: {resp.status}")
            if "text/html" not in resp.headers.get("content-type", ""):
                raise Exception("invalid content-type in response")
            text = await resp.text()
            return text
        
    async def crawl_page(self, current_url=None):
        if self.should_stop:
            return

        if current_url == None:
            current_url = self.base_url

        parsed_current = urlparse(current_url)
        if parsed_current.netloc != self.base_domain:
            return
        
        normalized_current = normalize_url(current_url)
        is_first = await self.add_page_visit(normalized_current)
        if not is_first:
            return

        async with self.semaphore:

            print(f"Fetching HTML from {current_url}")
            try:
                body = await self.get_html(current_url)
            except Exception as e:
                print(f"Error: {e}")
                return

            page = extract_page_data(body, current_url)
            async with self.lock:
                self.page_data[normalized_current] = page

            next_urls = get_urls_from_html(body, self.base_url)

            if self.should_stop:
                return
                
            tasks = []
            for link in next_urls:
                task = asyncio.create_task(self.crawl_page(link))
                tasks.append(task)
                self.all_tasks.add(task)

                    
            if tasks:
                try:
                    await asyncio.gather(*tasks, return_exceptions=True)
                finally:
                    for task in tasks:
                        self.all_tasks.discard(task)
    
    async def crawl(self):
        await self.crawl_page()
        return self.page_data