from typing import List
from datasets import Dataset
from transformers import GPT2Tokenizer, GPT2LMHeadModel
from transformers import Trainer, TrainingArguments
import nltk
from nltk.tokenize import sent_tokenize


nltk.download("punkt_tab")
nltk.download("punkt")
nltk.download("wordnet")
nltk.download("omw")


def load_and_split_sentences(train_file_path: str) -> List[str]:
    with open(train_file_path, "r") as file:
        text = file.read()

    sentences = sent_tokenize(text)
    return sentences


def train(
    train_file_path: str,
    output_dir: str,
    per_device_train_batch_size: int,
    num_train_epochs: float,
):
    tokenizer = GPT2Tokenizer.from_pretrained("openai-community/gpt2-medium")
    model = GPT2LMHeadModel.from_pretrained("openai-community/gpt2-medium")
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
        encoding["labels"] = encoding["input_ids"].clone()
        encoding["labels"][encoding["input_ids"] == tokenizer.pad_token_id] = -100
        return encoding

    sentences = load_and_split_sentences(train_file_path)
    data_dict = {"text": sentences}
    dataset = Dataset.from_dict(data_dict)
    tokenized_dataset = dataset.map(tokenize_sentences, batched=True)

    training_args = TrainingArguments(
        output_dir=output_dir,
        overwrite_output_dir=True,
        per_device_train_batch_size=per_device_train_batch_size,
        learning_rate=1e-4,
        weight_decay=0.01,
        num_train_epochs=num_train_epochs,
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
    train_file_path = "./data/bible.txt"
    output_dir = "./output"
    per_device_train_batch_size = 1
    num_train_epochs = 16

    train(
        train_file_path=train_file_path,
        output_dir=output_dir,
        per_device_train_batch_size=per_device_train_batch_size,
        num_train_epochs=num_train_epochs,
    )


if __name__ == "__main__":
    main()