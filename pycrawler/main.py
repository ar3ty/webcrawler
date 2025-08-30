import sys, asyncio, csv
from crawler import crawl_site_async

def write_csv_report(page_data, filename="report.csv"):
    with open(filename, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, ["page_url", "h1", "first_paragraph", "outgoing_link_urls", "image_urls"])
        writer.writeheader()
        for page in page_data.values():
            if page is None:
                continue
            writer.writerow({"page_url": page["url"], 
                             "h1": page["h1"], 
                             "first_paragraph": page["first_paragraph"], 
                             "outgoing_link_urls": ";".join(page["outgoing_links"]), 
                             "image_urls": ";".join(page["image_urls"])})

async def main_async():
    args = sys.argv
    if len(args) != 4:
        print("usage: uv run main.py URL max_concurrency max_pages")
        sys.exit(1)

    base_url = args[1]

    if not args[2].isdigit():
        print("max_concurrency must be integer")
        sys.exit(1)
    if not args[3].isdigit():
        print("max_pages must be integer")
        sys.exit(1)



    max_concurrency = int(args[2])
    max_pages = int(args[3])

    print(f"Crawling of: {base_url}...")

    
    page_data = await crawl_site_async(base_url, max_concurrency, max_pages)

    write_csv_report(page_data)

    sys.exit(0)

if __name__ == "__main__":
    asyncio.run(main_async())