// frontend/src/App.tsx

import React, { useState, useEffect } from 'react';
import './App.css';

// --- ИНТЕРФЕЙСЫ ТИПОВ ДАННЫХ ---
// Определяют, как выглядят данные, которые мы получаем от бэкенда.

interface AbilityScores {
  strength: number;
  dexterity: number;
  constitution: number;
  intelligence: number;
  wisdom: number;
  charisma: number;
}

interface CharacterSheet {
  name: string;
  abilityScores: AbilityScores;
  abilityModifiers: AbilityScores;
  skills: { [key: string]: number }; // Объект (карта), где ключ - название навыка, а значение - его числовое значение.
  skillMap: { [key: string]: string }; // Карта "Название навыка" -> "Название характеристики".
}

// --- КОМПОНЕНТЫ ---

// 1. Компонент для отображения одной характеристики
interface AbilityProps {
  name: string;
  score: number;
  modifier: number;
  onScoreChange: (newScore: number) => void; // Функция, которая вызывается при изменении значения в поле ввода
}

function Ability({ name, score, modifier, onScoreChange }: AbilityProps) {
  // Форматирует число модификатора в строку (например, 3 -> "+3", -1 -> "-1")
  const formatModifier = (mod: number) => (mod >= 0 ? `+${mod}` : mod.toString());

  // Обработчик события изменения в поле ввода
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    // Превращаем текст из поля в число. Если текст не является числом, используем 0.
    const newScore = parseInt(event.target.value, 10) || 0;
    onScoreChange(newScore);
  };

  return (
    <div className="ability">
      <div className="ability-name">{name}</div>
      <input
        type="number"
        className="ability-score-input"
        value={score}
        onChange={handleChange}
      />
      <div className="ability-modifier">({formatModifier(modifier)})</div>
    </div>
  );
}

// 2. Компонент для отображения списка всех навыков
interface SkillListProps {
  skills: { [key: string]: number };
  skillMap: { [key: string]: string };
}

function SkillList({ skills, skillMap }: SkillListProps) {
  // Получаем и сортируем названия навыков по алфавиту для стабильного порядка отображения
  const sortedSkillNames = Object.keys(skills).sort();
  
  // Функция для получения сокращенного названия характеристики (Strength -> Сил)
  const getAbilityAbbreviation = (abilityName: string) => {
    switch (abilityName) {
      case "Strength": return "Сил";
      case "Dexterity": return "Лов";
      case "Constitution": return "Тел";
      case "Intelligence": return "Инт";
      case "Wisdom": return "Мдр";
      case "Charisma": return "Хар";
      default: return "";
    }
  };

  return (
    <div className="skills-container">
      <h2>Навыки</h2>
      {sortedSkillNames.map(skillName => (
        <div className="skill-item" key={skillName}>
          <span className="skill-value">
            {skills[skillName] >= 0 ? `+${skills[skillName]}` : skills[skillName]}
          </span>
          <span className="skill-name">
            {skillName}
            <span className="skill-ability"> ({getAbilityAbbreviation(skillMap[skillName])})</span>
          </span>
        </div>
      ))}
    </div>
  );
}

// 3. Основной компонент приложения
function App() {
  // Состояние (state) для хранения всего листа персонажа. null - пока данные не загружены.
  const [characterSheet, setCharacterSheet] = useState<CharacterSheet | null>(null);
  
  // `useEffect` с пустым массивом [] запускает код один раз, после первого рендера компонента.
  // Идеально для первоначальной загрузки данных.
  useEffect(() => {
    fetch('http://localhost:8080/api/character')
      .then(response => response.json())
      .then((data: CharacterSheet) => setCharacterSheet(data)) // Сохраняем полученные данные в состояние
      .catch(error => console.error("Ошибка при загрузке данных:", error));
  }, []);
  
  // Функция для обновления состояния при изменении значения в поле ввода характеристики
  const handleAbilityChange = (abilityName: keyof AbilityScores, newScore: number) => {
    if (!characterSheet) return; // Если данных еще нет, ничего не делаем

    // Создаем новый, обновленный объект листа персонажа.
    // Важно создавать новые объекты, а не мутировать старые, чтобы React корректно отслеживал изменения.
    const updatedSheet = {
      ...characterSheet,
      abilityScores: {
        ...characterSheet.abilityScores,
        [abilityName]: newScore, // Динамически обновляем нужное поле (например, 'strength')
      },
    };
    setCharacterSheet(updatedSheet); // Обновляем состояние, что вызовет перерисовку интерфейса
  };

  // Функция для отправки изменений на сервер
  const handleSaveChanges = () => {
    if (!characterSheet) return;

    fetch('http://localhost:8080/api/character', {
      method: 'POST', // Используем метод POST для отправки данных
      headers: {
        'Content-Type': 'application/json', // Сообщаем серверу, что отправляем данные в формате JSON
      },
      body: JSON.stringify({ // Превращаем наш объект в JSON-строку для отправки
        name: characterSheet.name,
        abilityScores: characterSheet.abilityScores,
      }),
    })
      .then(response => response.json())
      .then((updatedSheetFromServer: CharacterSheet) => {
        // Когда сервер ответил, обновляем наше состояние его данными.
        // Это гарантирует, что мы отображаем актуальные, пересчитанные сервером значения (модификаторы, навыки).
        setCharacterSheet(updatedSheetFromServer);
        alert('Персонаж сохранен!');
      })
      .catch(error => console.error("Ошибка при сохранении:", error));
  };

  // Пока данные загружаются, показываем сообщение
  if (!characterSheet) {
    return <div className="App-header">Загрузка данных персонажа...</div>;
  }
  
  // Когда данные загружены, отрисовываем полный интерфейс
  return (
    <div className="App">
      <header className="App-header">
        <h1>{characterSheet.name}</h1>
        <div className="main-content">
          <div className="abilities-wrapper">
            <div className="abilities-container">
              <Ability name="Сила" score={characterSheet.abilityScores.strength} modifier={characterSheet.abilityModifiers.strength} onScoreChange={score => handleAbilityChange('strength', score)} />
              <Ability name="Ловкость" score={characterSheet.abilityScores.dexterity} modifier={characterSheet.abilityModifiers.dexterity} onScoreChange={score => handleAbilityChange('dexterity', score)} />
              <Ability name="Телосложение" score={characterSheet.abilityScores.constitution} modifier={characterSheet.abilityModifiers.constitution} onScoreChange={score => handleAbilityChange('constitution', score)} />
              <Ability name="Интеллект" score={characterSheet.abilityScores.intelligence} modifier={characterSheet.abilityModifiers.intelligence} onScoreChange={score => handleAbilityChange('intelligence', score)} />
              <Ability name="Мудрость" score={characterSheet.abilityScores.wisdom} modifier={characterSheet.abilityModifiers.wisdom} onScoreChange={score => handleAbilityChange('wisdom', score)} />
              <Ability name="Харизма" score={characterSheet.abilityScores.charisma} modifier={characterSheet.abilityModifiers.charisma} onScoreChange={score => handleAbilityChange('charisma', score)} />
            </div>
            <button className="save-button" onClick={handleSaveChanges}>
              Сохранить изменения
            </button>
          </div>
          <SkillList skills={characterSheet.skills} skillMap={characterSheet.skillMap} />
        </div>
      </header>
    </div>
  );
}

export default App;