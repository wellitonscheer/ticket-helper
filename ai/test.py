from transformers import (
    PreTrainedModel,
    AutoTokenizer,
    AutoModelForCausalLM
)
import torch

def load_model() -> PreTrainedModel:
    model = AutoModelForCausalLM.from_pretrained("meta-llama/Llama-3.2-1B")
    model.to("cuda")  # Move the model to GPU
    return model

def load_tokenizer():
    tokenizer = AutoTokenizer.from_pretrained("meta-llama/Llama-3.2-1B")
    return tokenizer

def generate_text(sequence: str) -> str:
    model = load_model()
    tokenizer = load_tokenizer()
    
    # Encode the input sequence and move to GPU
    encoding = tokenizer(sequence, return_tensors="pt").to("cuda")
    
    outputs = model.generate(
        encoding["input_ids"],
        attention_mask=encoding["attention_mask"],
        do_sample=True,
        max_length=150,
        eos_token_id=model.config.eos_token_id,
        early_stopping=True,
        pad_token_id=model.config.eos_token_id,
        top_p=0.95,
        temperature=0.5,
        num_beams=3,
    )

    return tokenizer.decode(outputs[0], skip_special_tokens=True)

def main():
    sequence = "who is the president of the united states?"
    result = generate_text(sequence)
    print(result)

if __name__ == "__main__":
    main()