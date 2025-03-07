
window.onload = function() {
  const el = document.getElementById("auth-redir");

  el.addEventListener('click', () => {
    window.location.href = window.location.origin + `/auth?user=123`;
  });
};


function redirectToCustomURL() {
  const userInput = document.getElementById("userInput");
  if (userInput.value === "") {
    return;
  }
  window.location.href = window.location.origin + `/auth?user=${userInput.value}`;
}
