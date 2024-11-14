from typing import List, Dict
from datasets import Dataset, DatasetDict
from transformers import GPT2Tokenizer, GPT2LMHeadModel
from transformers import Trainer, TrainingArguments
import nltk
from nltk.tokenize import sent_tokenize
import numpy as np
import torch


nltk.download("punkt")


def load_and_split_sentences(train_file_path: str) -> List[str]:
    with open(train_file_path, "r") as file:
        text = file.read()

    sentences = sent_tokenize(text, "portuguese")
    return sentences


def train(
    train_file_path: str,
    output_dir: str,
    per_device_train_batch_size: int,
    num_train_epochs: float,
):
    # tokenizer = GPT2Tokenizer.from_pretrained("openai-community/gpt2-medium")
    # model = GPT2LMHeadModel.from_pretrained("openai-community/gpt2-medium")
    tokenizer = GPT2Tokenizer.from_pretrained("Locutusque/gpt2-conversational-or-qa")
    model = GPT2LMHeadModel.from_pretrained("Locutusque/gpt2-conversational-or-qa")
    model.to("cuda")

    if tokenizer.pad_token is None:
        tokenizer.add_special_tokens({"pad_token": "[PAD]"})
        model.resize_token_embeddings(len(tokenizer))

    def tokenize_sentences(sentences):
        encoding = tokenizer(
            sentences["text"],
            return_tensors="pt",
            padding="max_length",
            truncation=True,
        )
        encoding.to("cuda")
        encoding["labels"] = encoding["input_ids"].clone()
        encoding["labels"][encoding["input_ids"] == tokenizer.pad_token_id] = -100
        return encoding

    sentences = load_and_split_sentences(train_file_path)
    data_dict = {"text": sentences}
    dataset = Dataset.from_dict(data_dict)

    # Split into training and validation sets
    dataset = dataset.train_test_split(test_size=0.1)
    tokenized_datasets = dataset.map(tokenize_sentences, batched=True)

    # Define the training arguments with early stopping enabled
    training_args = TrainingArguments(
        output_dir=output_dir,
        overwrite_output_dir=True,
        per_device_train_batch_size=per_device_train_batch_size,
        learning_rate=1e-5,
        weight_decay=0.1,
        num_train_epochs=num_train_epochs,
        eval_strategy="epoch",
        save_strategy="epoch",
        load_best_model_at_end=True,  # Loads the best model at the end based on validation loss
        metric_for_best_model="eval_loss",  # Use validation loss as the metric for early stopping
        greater_is_better=False,  # We want the loss to decrease
    )

    # Define a function to compute metrics
    def compute_metrics(eval_pred):
        logits, labels = eval_pred
        predictions = np.argmax(logits, axis=-1)
        # Calculate loss only where labels are not -100 (ignored tokens)
        loss_fct = torch.nn.CrossEntropyLoss(ignore_index=-100)
        loss = loss_fct(
            torch.tensor(logits, dtype=torch.float32), 
            torch.tensor(labels, dtype=torch.long)
        )
        return {"eval_loss": loss.item()}

    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=tokenized_datasets["train"],
        eval_dataset=tokenized_datasets["test"],
        compute_metrics=compute_metrics,
    )

    trainer.train()

    tokenizer.save_pretrained(output_dir)
    model.save_pretrained(output_dir)
    trainer.save_model()


def main():
    train_file_path = "./get_data/data/output.txt"
    output_dir = "./output"
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
