chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  console.log({ sender, sendResponse });
  if (message.action === "getDivResponseMessage") {
    const div = document.querySelector(
      'div[placeholder="Digite sua resposta aqui. Use as respostas predefinidas a partir do menu suspenso"]'
    );
    if (div) {
      const inner = div.innerHTML;
      console.log("Child Elements:", inner);

      sendResponse({ inner });
    } else {
      console.log("Div not found.");
      sendResponse({ error: "Div not found" });
    }
  }

  return true;
});
