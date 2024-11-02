document.getElementById("correctMessage").addEventListener("click", myAlert);

function myAlert() {
  console.log("correcting!");
  chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
    chrome.tabs.sendMessage(
      tabs[0].id,
      { action: "getDivResponseMessage" },
      (response) => {
        if (response.error) {
          console.log(response.error);
        } else {
          console.log("Elements in Div:", response.inner);
        }
      }
    );
  });
}
