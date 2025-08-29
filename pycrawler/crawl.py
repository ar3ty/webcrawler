import urllib.parse as up
from bs4 import BeautifulSoup as bs

def normalize_url(url: str) -> str:
    parsedURL = up.urlparse(url)
    combined = parsedURL.netloc + parsedURL.path
    return combined.rstrip('/').lower()

def get_h1_from_html(html):
    soup = bs(html, 'html.parser')
    h1_tag = soup.find("h1")
    return h1_tag.get_text(strip=True) if h1_tag else ""

def get_first_paragraph_from_html(html):
    soup = bs(html, 'html.parser')
    main = soup.find('main')
    if main:
        first_p = main.find('p')
    else:
        first_p = soup.find('p')
    return first_p.get_text(strip=True) if first_p else ""

def get_urls_from_html(html, base_url):
    result = []
    soup = bs(html, 'html.parser')
    urls = soup.find_all('a')

    for url in urls:
        if href := url.get('href'):
            try:
                absURL = up.urljoin(base_url, href)
                result.append(absURL)
            except Exception as e:
                print(f"{str(e)}: {href}")

    return result

def get_images_from_html(html, base_url):
    result = []
    soup = bs(html, 'html.parser')
    urls = soup.find_all('img')

    for url in urls:
        if src := url.get('src'):
            try:
                absURL = up.urljoin(base_url, src)
                result.append(absURL)
            except Exception as e:
                print(f"{str(e)}: {src}")

    return result

def extract_page_data(html, page_url):
    return {
        "url": page_url,
        "h1": get_h1_from_html(html),
        "first_paragraph": get_first_paragraph_from_html(html),
        "outgoing_links": get_urls_from_html(html, page_url),
        "image_urls": get_images_from_html(html, page_url) 
    }