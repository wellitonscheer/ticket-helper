from typing import List, Dict
from datasets import Dataset
from transformers import BertTokenizer, BertForQuestionAnswering, Trainer, TrainingArguments
import json
import torch

def answer_question(model, tokenizer, context: str) -> str:
    inputs = tokenizer(context, return_tensors="pt", max_length=512, truncation=True)
    with torch.no_grad():
        outputs = model(**inputs)
        start_logits = outputs.start_logits
        end_logits = outputs.end_logits

        start_index = torch.argmax(start_logits)
        end_index = torch.argmax(end_logits) + 1  # Add 1 to include end position

    answer = tokenizer.convert_tokens_to_string(
        tokenizer.convert_ids_to_tokens(inputs['input_ids'][0][start_index:end_index])
    )
    return answer


def main():
    train_file_path = "./data/output-squad.json"  # Should be in SQuAD-style JSON format
    output_dir = "./output-pt"
    per_device_train_batch_size = 1
    num_train_epochs = 8

    # Load the trained model and tokenizer
    tokenizer = BertTokenizer.from_pretrained(output_dir)
    model = BertForQuestionAnswering.from_pretrained(output_dir)

    # Example question answering
    context = "criar um evento no sistema de bolsas para o cadastro socioeconomico  vinculado a portaria anexa."

    answer = answer_question(model, tokenizer, context,)
    print("Answer:", answer)


if __name__ == "__main__":
    main()