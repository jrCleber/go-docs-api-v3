if (window.location.pathname === '/webhook/v1.0.0/identity-change') {
  window.addEventListener('DOMContentLoaded', (event) => {
    const element = document.querySelector('.tabs-container .openapi-tabs__code-container');
    if (element) {
      element.parentNode.removeChild(element);
    }
  });
}