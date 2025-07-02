import os

INPUT_FILE = "found_urls.txt"
IMAGE_DIR = "images"
OUTPUT_FILE = "README.md"

header = """\
scrape every fuckin BMA traffic camera

I don’t get their fuckass camera IDs, so I'm  just gonna brute-force that shit.

these are photos:
"""

table_header = "| Image | URL |\n|-------|-----|"

def get_filename_from_url(url):
    return url.strip().split("/")[-1]

def generate_row(image_filename, url):
    return f"| ![img](images/{image_filename}) | [{url}]({url}) |"

def main():
    if not os.path.exists(INPUT_FILE):
        print("No found_urls.txt found.")
        return

    rows = []
    with open(INPUT_FILE, "r") as f:
        for line in f:
            url = line.strip()
            if not url:
                continue
            filename = get_filename_from_url(url)
            local_path = os.path.join(IMAGE_DIR, filename)
            if os.path.exists(local_path):
                rows.append(generate_row(filename, url))

    with open(OUTPUT_FILE, "w") as out:
        out.write(header + "\n")
        out.write(table_header + "\n")
        for row in rows:
            out.write(row + "\n")

    print(f"README.md generated with {len(rows)} images.")

if __name__ == "__main__":
    main()
