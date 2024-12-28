import json
import re

def remove_html_tags(text):
    clean_text = re.sub(r'<.*?>', '', text)
    return clean_text

transformed_data = {}

with open('inputs/os_ticket_raw.json', 'r', encoding='utf-8') as file:
    data = json.load(file)

    for item in data["data"]:
        if "Recebemos sua solicitação, assim que possível lhe retornamos" in item["body"]:
            continue
        if "Seu feedback é muito importante" in item["body"]:
            continue
        if "A solicitação foi encerrada" in item["body"]:
            continue
        if "Task closed" in item["body"]:
            continue
        if "fechado automaticamente pela ausência de retorno" in item["body"]:
            continue

        ticket_id = str(item["ticket_id"])

        if ticket_id not in transformed_data:
            transformed_data[ticket_id] = []
        
        clean_body = remove_html_tags(item["body"])
        
        transformed_data[ticket_id].append({
            "type": item["type"],
            "ordem": item["ordem"],
            "subject": item["subject"],
            "poster": item["poster"],
            "body": clean_body
        })

for ticket_id in transformed_data:
    transformed_data[ticket_id].sort(key=lambda x: x["ordem"])

with open('outputs/id_list.json', 'w', encoding='utf-8') as file:
    json.dump(transformed_data, file, ensure_ascii=False, indent=2)

print("Transformation complete! The output is saved in 'output.json'.")