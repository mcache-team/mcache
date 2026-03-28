// Tab switching
document.querySelectorAll('.tab').forEach(tab => {
  tab.addEventListener('click', () => {
    const target = tab.dataset.tab;
    const parent = tab.closest('.section, .section-dark');

    parent.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    parent.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

    tab.classList.add('active');
    const content = parent.querySelector(`#tab-${target}`);
    if (content) content.classList.add('active');
  });
});

// Highlight active nav link on scroll
const sections = document.querySelectorAll('section[id]');
const navLinks = document.querySelectorAll('.nav-links a[href^="#"]');

const observer = new IntersectionObserver(entries => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      navLinks.forEach(a => a.classList.remove('active-link'));
      const link = document.querySelector(`.nav-links a[href="#${entry.target.id}"]`);
      if (link) link.classList.add('active-link');
    }
  });
}, { rootMargin: '-40% 0px -55% 0px' });

sections.forEach(s => observer.observe(s));
