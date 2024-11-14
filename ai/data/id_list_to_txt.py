import json
import re

# Read the JSON data
with open('id_list.json', 'r') as json_file:
    data = json.load(json_file)

result_lines = {}
result_lines["data"] = []
for ticket_id, ticket_data in data.items():
    # Loop through the 'body' of each item in the ticket data
    # combined_bodies = ' '.join([body["body"] for body in ticket_data])
    # result_lines.append(combined_bodies)
    
    for idx, item in enumerate(ticket_data):
        if idx == len(ticket_data) - 1:
            break
        combined_bodies = ' - '.join([ticket_data[i]['body'] for i in range(idx + 1)])
        combined_bodies = re.sub(r'\s+', ' ', combined_bodies).strip()
        result_lines["data"].append({
            "context": combined_bodies,
            "answer": f'{ticket_data[idx + 1]["body"]}'
            }
        )

    # for idx, item in enumerate(ticket_data):
    #     if idx == len(ticket_data) - 1:
    #         break
    #     print(idx, item)
    #     combined_bodies = ' '.join([ticket_data[i]['body'] for i in range(idx + 1)])
    #     print(combined_bodies)
    #     result_lines.append(f'me de a resposta: "{combined_bodies}"')
    #     result_lines.append(f'resposta: {ticket_data[idx + 1]["body"]}')

# Write the result to a text file
with open('output-squad.json', 'w', encoding='utf-8') as txt_file:
    json.dump(result_lines, txt_file, ensure_ascii=False, indent=2)
    # txt_file.write('\n'.join(result_lines))