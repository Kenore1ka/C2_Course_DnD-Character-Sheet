// frontend/src/App.tsx

import { useState, useEffect } from 'react';
import './App.css'; // Оставляем стили для красоты

function App() {
  // Создаем "состояние" для хранения ответа от сервера
  // status - переменная, setStatus - функция для ее изменения
  const [status, setStatus] = useState('Загрузка...');

  // useEffect - это специальная функция в React,
  // которая выполняет код один раз после отрисовки компонента.
  // Идеально подходит для запросов к серверу.
  useEffect(() => {
    // Используем встроенную в браузер функцию fetch для отправки запроса
    fetch('http://localhost:8080/api/health') // <-- Наш URL бэкенда
      .then(response => response.json()) // Превращаем ответ в JSON
      .then(data => {
        // Когда данные получены, обновляем наше состояние
        setStatus(data.status); // data будет { status: "ok" }
      })
      .catch(error => {
        // Если произошла ошибка (например, сервер выключен), сообщаем об этом
        console.error("Ошибка при запросе к бэкенду:", error);
        setStatus('Ошибка!');
      });
  }, []); // Пустой массив [] означает, что эффект выполнится только один раз

  return (
    <div className="App">
      <header className="App-header">
        <h1>Лист персонажа D&D</h1>
        <p>
          Статус сервера: <strong>{status}</strong>
        </p>
      </header>
    </div>
  );
}

export default App;