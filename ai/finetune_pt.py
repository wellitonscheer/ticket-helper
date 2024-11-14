from typing import List, Dict
from datasets import Dataset
from transformers import BertTokenizer, BertForQuestionAnswering
from transformers import Trainer, TrainingArguments
import json


def load_qa_data(train_file_path: str) -> List[Dict[str, str]]:
    with open(train_file_path, "r") as file:
        data = json.load(file)
    return data["data"]


def train(
    train_file_path: str,
    output_dir: str,
    per_device_train_batch_size: int,
    num_train_epochs: float,
):
    tokenizer = BertTokenizer.from_pretrained("neuralmind/bert-base-portuguese-cased")
    model = BertForQuestionAnswering.from_pretrained("neuralmind/bert-base-portuguese-cased")
    model.to("cuda")

    def preprocess_function(examples):
        contexts = examples["context"]
        answers = examples["answer"]

        inputs = tokenizer(
            contexts,
            truncation=True,
            padding="max_length",
            return_tensors="pt",
            max_length=512
        )

        start_positions = []
        end_positions = []
        for i, answer in enumerate(answers):
            start_idx = contexts[i].find(answer)
            end_idx = start_idx + len(answer)
            start_positions.append(start_idx)
            end_positions.append(end_idx)

        inputs["start_positions"] = start_positions
        inputs["end_positions"] = end_positions

        return inputs

    qa_data = load_qa_data(train_file_path)
    dataset = Dataset.from_dict({"context": [qa["context"] for qa in qa_data], 
                                 "answer": [qa["answer"] for qa in qa_data]})
    tokenized_dataset = dataset.map(preprocess_function, batched=True)

    training_args = TrainingArguments(
        output_dir=output_dir,
        overwrite_output_dir=True,
        per_device_train_batch_size=per_device_train_batch_size,
        learning_rate=1e-4,
        weight_decay=0.01,
        num_train_epochs=num_train_epochs,
        save_total_limit=2
    )

    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=tokenized_dataset,
    )

    trainer.train()

    tokenizer.save_pretrained(output_dir)
    model.save_pretrained(output_dir)
    trainer.save_model()


def main():
    train_file_path = "./data/output-squad.json"
    output_dir = "./output-pt"
    per_device_train_batch_size = 1
    num_train_epochs = 8

    train(
        train_file_path=train_file_path,
        output_dir=output_dir,
        per_device_train_batch_size=per_device_train_batch_size,
        num_train_epochs=num_train_epochs,
    )


if __name__ == "__main__":
    main()
