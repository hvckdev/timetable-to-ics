const input = document.querySelector('#group-input');
const options = document.querySelector('#group-options');
const availability = document.querySelector('#availability');
const hint = document.querySelector('#field-hint');
const clearButton = document.querySelector('#clear-button');
const appleButton = document.querySelector('#apple-button');
const googleButton = document.querySelector('#google-button');
const downloadButton = document.querySelector('#download-button');
const form = document.querySelector('#calendar-form');
const notice = document.querySelector('#notice');

let groups = [];
let visibleGroups = [];
let activeIndex = -1;

const normalize = (value) => value.trim().toLocaleLowerCase('ru-RU').replaceAll('ё', 'е');

function selectedGroup() {
  const value = normalize(input.value);
  return groups.find((group) => normalize(group) === value) || '';
}

function calendarURL(group) {
  const url = new URL('/calendar', window.location.origin);
  url.searchParams.set('group', group);
  return url;
}

function setActionsEnabled(enabled) {
  appleButton.disabled = !enabled;
  googleButton.disabled = !enabled;
  downloadButton.disabled = !enabled;
}

function updateSelectionState() {
  const valid = Boolean(selectedGroup());
  setActionsEnabled(valid);
  clearButton.hidden = input.value.length === 0;
  hint.classList.remove('invalid');
  hint.textContent = valid
    ? 'Группа найдена — выберите календарь ниже'
    : 'Можно начать вводить название — покажем совпадения';
}

function closeOptions() {
  options.hidden = true;
  input.setAttribute('aria-expanded', 'false');
  input.removeAttribute('aria-activedescendant');
  activeIndex = -1;
}

function chooseGroup(group) {
  input.value = group;
  closeOptions();
  updateSelectionState();
  input.focus();
}

function renderOptions() {
  const query = normalize(input.value);
  visibleGroups = groups
    .filter((group) => !query || normalize(group).includes(query))
    .slice(0, 9);

  options.replaceChildren();
  activeIndex = -1;

  if (visibleGroups.length === 0) {
    closeOptions();
    if (query) {
      hint.textContent = 'Такой группы нет в актуальном расписании';
      hint.classList.add('invalid');
    }
    return;
  }

  visibleGroups.forEach((group, index) => {
    const option = document.createElement('button');
    option.type = 'button';
    option.className = 'option';
    option.id = `group-option-${index}`;
    option.setAttribute('role', 'option');

    const name = document.createElement('strong');
    name.textContent = group;
    const marker = document.createElement('span');
    marker.textContent = 'выбрать';
    option.append(name, marker);
    option.addEventListener('mousedown', (event) => event.preventDefault());
    option.addEventListener('click', () => chooseGroup(group));
    options.append(option);
  });

  options.hidden = false;
  input.setAttribute('aria-expanded', 'true');
}

function moveActive(direction) {
  if (options.hidden || visibleGroups.length === 0) {
    renderOptions();
  }
  if (visibleGroups.length === 0) return;

  activeIndex = (activeIndex + direction + visibleGroups.length) % visibleGroups.length;
  options.querySelectorAll('.option').forEach((option, index) => {
    option.classList.toggle('active', index === activeIndex);
    option.setAttribute('aria-selected', String(index === activeIndex));
  });
  const active = document.querySelector(`#group-option-${activeIndex}`);
  input.setAttribute('aria-activedescendant', active.id);
  active.scrollIntoView({ block: 'nearest' });
}

async function loadGroups() {
  try {
    const response = await fetch('/api/groups', { headers: { Accept: 'application/json' } });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const payload = await response.json();
    groups = Array.isArray(payload.groups) ? payload.groups : [];
    if (groups.length === 0) throw new Error('empty group list');

    input.disabled = false;
    availability.classList.add('ready');
    availability.querySelector('span:last-child').textContent = `${groups.length} групп доступно`;
  } catch (error) {
    console.error('Cannot load groups:', error);
    availability.classList.add('error');
    availability.querySelector('span:last-child').textContent = 'Не удалось загрузить';
    hint.textContent = 'Обновите страницу или попробуйте немного позже';
    hint.classList.add('invalid');
  }
}

input.addEventListener('input', () => {
  updateSelectionState();
  renderOptions();
});
input.addEventListener('focus', renderOptions);
input.addEventListener('blur', () => window.setTimeout(closeOptions, 120));
input.addEventListener('keydown', (event) => {
  if (event.key === 'ArrowDown') {
    event.preventDefault();
    moveActive(1);
  } else if (event.key === 'ArrowUp') {
    event.preventDefault();
    moveActive(-1);
  } else if (event.key === 'Enter' && activeIndex >= 0) {
    event.preventDefault();
    chooseGroup(visibleGroups[activeIndex]);
  } else if (event.key === 'Escape') {
    closeOptions();
  }
});

clearButton.addEventListener('click', () => {
  input.value = '';
  updateSelectionState();
  renderOptions();
  input.focus();
});

appleButton.addEventListener('click', () => {
  const group = selectedGroup();
  if (!group) return;
  const subscriptionURL = calendarURL(group).toString().replace(/^https?:/, 'webcal:');
  window.location.href = subscriptionURL;
});

googleButton.addEventListener('click', () => {
  const group = selectedGroup();
  if (!group) return;

  if (['localhost', '127.0.0.1', '::1'].includes(window.location.hostname)) {
    notice.hidden = false;
    notice.textContent = 'Google Calendar не может загрузить календарь с localhost. Откройте опубликованную версию сервиса или скачайте .ics.';
    return;
  }

  notice.hidden = true;
  const googleURL = new URL('https://calendar.google.com/calendar/render');
  googleURL.searchParams.set('cid', calendarURL(group).toString());
  window.open(googleURL.toString(), '_blank', 'noopener,noreferrer');
});

form.addEventListener('submit', (event) => {
  event.preventDefault();
  const group = selectedGroup();
  if (!group) return;
  window.location.href = calendarURL(group).toString();
});

loadGroups();
