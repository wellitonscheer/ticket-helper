import os

data_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "data")
output_file = os.path.join(data_dir, "data.txt")

if os.path.exists(output_file):
    os.remove(output_file)

page_number = 1

while True:
    file_path = os.path.join(data_dir, f"page_data_{page_number}.txt")
    
    if os.path.exists(file_path):
        with open(file_path, "r", encoding="utf-8") as file:
            content = file.read()
        
        with open(output_file, "a", encoding="utf-8") as output:
            output.write(content)
            output.write("\n")
        
        print(f"Added content from {file_path} to {output_file}")
    else:
        print(f"{file_path} does not exist. Stopping the loop.")
        break

    page_number += 1
