from transformers import (
    PreTrainedModel,
    GPT2LMHeadModel,
    GPT2Tokenizer,
    BertTokenizer,
    BertForQuestionAnswering
)


def load_model(model_path: str) -> PreTrainedModel:
    model = GPT2LMHeadModel.from_pretrained(model_path)
    model.to("cuda")
    return model


def load_tokenizer(tokenizer_path: str):
    tokenizer = GPT2Tokenizer.from_pretrained(tokenizer_path)
    return tokenizer


def generate_text(sequence: str) -> str:
    model_path = "./output"
    model = load_model(model_path)
    tokenizer = load_tokenizer(model_path)

    encoding = tokenizer(
        sequence,
        return_tensors="pt",
        padding="max_length",
        truncation=True,
    ).to("cuda")

    outputs = model.generate(
        encoding["input_ids"],
        attention_mask=encoding["attention_mask"],
        do_sample=True,
        max_new_tokens=128,
        eos_token_id=model.config.eos_token_id,
        early_stopping=True,
        pad_token_id=model.config.eos_token_id,
        # top_k=50,
        top_p=0.95,
        # num_return_sequences=1,
        # no_repeat_ngram_size=2,
        temperature=0.9,
        num_beams=3,
    )

    return tokenizer.decode(outputs[0], skip_special_tokens=True)


def main():
    sequence = '''
    Olá, tudo bem? 
    Preciso de ajuda para baixar/encontrar o logos no IOS, não estou encontrado!
'''
    result = generate_text(sequence)
    print("result aaaaaaaaaaaaaaaaaaaaaaaaa\n")
    print(result)


if __name__ == "__main__":
    main()
