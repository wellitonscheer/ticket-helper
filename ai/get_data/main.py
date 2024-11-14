import requests
import logging
import os
import threading
from bs4 import BeautifulSoup

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s")

headers = {
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
    "Accept-Encoding": "gzip, deflate, br, zstd",
    "Accept-Language": "en-US,en;q=0.9,pt-BR;q=0.8,pt;q=0.7",
    "Cache-Control": "max-age=0",
    "Connection": "keep-alive",
    "DNT": "1",
    "Referer": "https://suporte.setrem.com.br/scp/login.php",
    "Sec-Fetch-Dest": "document",
    "Sec-Fetch-Mode": "navigate",
    "Sec-Fetch-Site": "same-origin",
    "Sec-Fetch-User": "?1",
    "Sec-GPC": "1",
    "Upgrade-Insecure-Requests": "1",
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
    "sec-ch-ua": '"Chromium";v="130", "Brave";v="130", "Not?A_Brand";v="99"',
    "sec-ch-ua-mobile": "?0",
    "sec-ch-ua-platform": '"Linux"'
}

cookies = {
    "OSTSESSID": "s61217en0dp549mj7ssnnbf8hb"
}

def fetch_and_save(thread_id, first_page, last_page):
    file_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), f"data/page_data_{thread_id}.txt")

    for page_id in range(first_page, last_page + 1):
        url = f"https://suporte.setrem.com.br/scp/tickets.php?id={page_id}"
        
        try:
            logging.info("Thread %s: Sending GET request to %s", thread_id, url)
            
            response = requests.get(url, headers=headers, cookies=cookies)
            
            if response.status_code == 200:
                logging.info("Thread %s: Page_id %d: Request successful! Status code: %d", thread_id, page_id, response.status_code)
                
                soup = BeautifulSoup(response.text, 'html.parser')
                
                thread_items_div = soup.find("div", {"id": "thread-items"})
                
                if thread_items_div:
                    with open(file_path, "w", encoding="utf-8") as file:
                        file.write(str(thread_items_div))
                    logging.info("Thread %s: Page_id %d: Extracted <div id='thread-items'> saved to %s", thread_id, page_id, file_path)
                else:
                    logging.warning("Thread %s: Page_id %d: The <div id='thread-items'> was not found in the response.", thread_id, page_id)
            
            else:
                logging.warning("Thread %s: Page_id %d: Request returned with status code: %d", thread_id, page_id, response.status_code)
        
        except requests.exceptions.RequestException as e:
            logging.error("Thread %s: Page_id %d: Request failed: %s", thread_id, page_id, e)

pages_per_thread = 700
last_search_page = 88
threads = []
amount_threads = 50
for thread_id in range(1, amount_threads + 1):
    thread_first_page = last_search_page
    last_search_page = last_search_page + pages_per_thread
    thread = threading.Thread(target=fetch_and_save, args=(thread_id,thread_first_page,last_search_page,))
    threads.append(thread)
    thread.start()

for thread in threads:
    thread.join()

logging.info("All threads completed.")