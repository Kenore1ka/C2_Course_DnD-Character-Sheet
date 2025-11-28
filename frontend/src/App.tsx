// frontend/src/App.tsx

import React, { useState, useEffect } from 'react';
import './App.css';

// --- ИНТЕРФЕЙСЫ ТИПОВ ДАННЫХ ---

interface HitPoints { current: number; max: number; }
interface AbilityScores { strength: number; dexterity: number; constitution: number; intelligence: number; wisdom: number; charisma: number; }
interface Item { id: number; name: string; type: string; description: string; }
interface CharacterItem { itemId: number; quantity: number; }
interface InventoryItem { item: Item; quantity: number; }

interface CharacterSheet {
  name: string; class: string; race: string; alignment: string; level: number; proficiencyBonus: number;
  hitPoints: HitPoints; armorClass: number; initiative: number; abilityScores: AbilityScores;
  abilityModifiers: AbilityScores; savingThrows: { [key: string]: number }; skills: { [key: string]: number };
  skillMap: { [key: string]: string }; skillProficiencies: string[]; savingThrowProficiencies: string[];
}

// --- КОМПОНЕНТЫ ---

// 1. Компонент для основной информации
interface CharacterInfoProps {
  name: string; characterClass: string; race: string; alignment: string;
  onInfoChange: (field: 'name' | 'class' | 'race' | 'alignment', value: string) => void;
}
function CharacterInfo({ name, characterClass, race, alignment, onInfoChange }: CharacterInfoProps) {
  return (
    <div className="character-info-container">
      <div className="info-item main-name"><input type="text" value={name} onChange={e => onInfoChange('name', e.target.value)} /><label>Имя персонажа</label></div>
      <div className="info-item"><input type="text" value={characterClass} onChange={e => onInfoChange('class', e.target.value)} /><label>Класс</label></div>
      <div className="info-item"><input type="text" value={race} onChange={e => onInfoChange('race', e.target.value)} /><label>Раса</label></div>
      <div className="info-item"><input type="text" value={alignment} onChange={e => onInfoChange('alignment', e.target.value)} /><label>Мировоззрение</label></div>
    </div>
  );
}

// 2. Компонент для одной характеристики
interface AbilityProps { name: string; score: number; modifier: number; onScoreChange: (newScore: number) => void; }
function Ability({ name, score, modifier, onScoreChange }: AbilityProps) {
  const formatModifier = (mod: number) => (mod >= 0 ? `+${mod}` : mod.toString());
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => { onScoreChange(parseInt(event.target.value, 10) || 0); };
  return (
    <div className="ability">
      <div className="ability-name">{name}</div>
      <input type="number" className="ability-score-input" value={score} onChange={handleChange} />
      <div className="ability-modifier">({formatModifier(modifier)})</div>
    </div>
  );
}

// 3. Компонент для списка Спасбросков
interface SavingThrowListProps { savingThrows: { [key: string]: number }; proficiencies: string[]; onProficiencyChange: (abilityName: string, isProficient: boolean) => void; }
function SavingThrowList({ savingThrows, proficiencies, onProficiencyChange }: SavingThrowListProps) {
  const abilityOrder: (keyof AbilityScores)[] = ["strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"];
  const abilityNamesRu: { [key: string]: string } = { strength: "Сила", dexterity: "Ловкость", constitution: "Телосложение", intelligence: "Интеллект", wisdom: "Мудрость", charisma: "Харизма" };
  return (
    <div className="saving-throws-container">
      <h3>Спасброски</h3>
      {abilityOrder.map(abilityKey => (
        <div className="skill-item" key={abilityKey}>
          <input type="checkbox" className="skill-proficiency-checkbox" checked={proficiencies.includes(abilityKey)} onChange={e => onProficiencyChange(abilityKey, e.target.checked)} />
          <span className="skill-value">{savingThrows[abilityKey] >= 0 ? `+${savingThrows[abilityKey]}` : savingThrows[abilityKey]}</span>
          <span className="skill-name">{abilityNamesRu[abilityKey]}</span>
        </div>
      ))}
    </div>
  );
}

// 4. Компонент для списка Навыков
interface SkillListProps { skills: { [key: string]: number }; skillMap: { [key: string]: string }; proficiencies: string[]; onProficiencyChange: (skillName: string, isProficient: boolean) => void; }
function SkillList({ skills, skillMap, proficiencies, onProficiencyChange }: SkillListProps) {
  const sortedSkillNames = Object.keys(skills).sort();
  const getAbilityAbbreviation = (abilityName: string) => abilityName.substring(0, 3).toUpperCase();
  return (
    <div className="skills-container">
      <h2>Навыки</h2>
      {sortedSkillNames.map(skillName => (
        <div className="skill-item" key={skillName}>
          <input type="checkbox" className="skill-proficiency-checkbox" checked={proficiencies.includes(skillName)} onChange={e => onProficiencyChange(skillName, e.target.checked)} />
          <span className="skill-value">{skills[skillName] >= 0 ? `+${skills[skillName]}` : skills[skillName]}</span>
          <span className="skill-name">{skillName} <span className="skill-ability">({getAbilityAbbreviation(skillMap[skillName])})</span></span>
        </div>
      ))}
    </div>
  );
}

// 5. Компонент для Инвентаря
interface InventoryProps { inventory: InventoryItem[]; onAddItem: (item: CharacterItem) => void; onDeleteItem: (itemId: number) => void; }
function Inventory({ inventory, onAddItem, onDeleteItem }: InventoryProps) {
  const [itemName, setItemName] = useState('');
  const [quantity, setQuantity] = useState(1);

  const handleAddItem = () => {
    let itemId = 0;
    if (itemName.toLowerCase().includes('меч')) itemId = 1;
    else if (itemName.toLowerCase().includes('броня')) itemId = 2;
    else if (itemName.toLowerCase().includes('зелье')) itemId = 3;
    else if (itemName.toLowerCase().includes('лук')) itemId = 4;
    else { alert('Неизвестный предмет. Введите "меч", "броня", "зелье" или "лук".'); return; }

    onAddItem({ itemId, quantity });
    setItemName('');
    setQuantity(1);
  };

  return (
    <div className="inventory-container">
      <h3>Инвентарь</h3>
      <ul className="inventory-list">
        {!inventory || inventory.length === 0 ? (
              <li>Пусто</li>
        ) : (
          inventory.map((invItem) => (
            <li key={invItem.item.id}>
              <div className="item-info">
                <span className="item-name">{invItem.item.name}</span>
                <span className="item-quantity">x{invItem.quantity}</span>
              </div>
              <button className="delete-item-button" onClick={() => onDeleteItem(invItem.item.id)}>&times;</button>
            </li>
          ))
        )}
      </ul>
      <div className="add-item-form">
        <input type="text" placeholder="Название предмета..." value={itemName} onChange={e => setItemName(e.target.value)} />
        <input type="number" min="1" value={quantity} onChange={e => setQuantity(parseInt(e.target.value, 10) || 1)} />
        <button onClick={handleAddItem}>Добавить</button>
      </div>
    </div>
  );
}

// 6. Компонент для Боевых параметров
interface CombatStatsProps { hitPoints: HitPoints; armorClass: number; initiative: number; onCurrentHpChange: (newHp: number) => void; }
function CombatStats({ hitPoints, armorClass, initiative, onCurrentHpChange }: CombatStatsProps) {
  return (
    <div className="combat-stats-container">
      <div className="combat-stat-item"><label>Класс доспеха</label><div className="combat-stat-value ac">{armorClass}</div></div>
      <div className="combat-stat-item"><label>Инициатива</label><div className="combat-stat-value initiative">{initiative >= 0 ? `+${initiative}` : initiative}</div></div>
      <div className="combat-stat-item hp"><label>Хиты</label><div className="hp-inputs"><input type="number" value={hitPoints.current} onChange={e => onCurrentHpChange(parseInt(e.target.value, 10) || 0)} /><span>/</span><span>{hitPoints.max}</span></div></div>
    </div>
  );
}

// 7. Компонент для Уровня и Бонуса Мастерства
interface VitalsProps { level: number; proficiencyBonus: number; onLevelChange: (newLevel: number) => void; }
function Vitals({ level, proficiencyBonus, onLevelChange }: VitalsProps) {
  return (
    <div className="vitals-container">
      <div className="vital-item"><label>Уровень</label><input type="number" value={level} onChange={e => onLevelChange(parseInt(e.target.value, 10) || 1)} min="1" max="20" /></div>
      <div className="vital-item"><label>Бонус мастерства</label><span className="vital-value">+{proficiencyBonus}</span></div>
    </div>
  );
}

// 8. Основной компонент приложения
function App() {
  const [characterSheet, setCharacterSheet] = useState<CharacterSheet | null>(null);
  const [inventory, setInventory] = useState<InventoryItem[]>([]);

  useEffect(() => {
    fetch('http://localhost:8080/api/character').then(res => res.json()).then(data => setCharacterSheet(data)).catch(console.error);
    fetch('http://localhost:8080/api/character/inventory').then(res => res.json()).then(data => setInventory(data || [])).catch(console.error);
  }, []);

  // --- ОБРАБОТЧИКИ ИЗМЕНЕНИЙ ---
  const handleInfoChange = (field: 'name' | 'class' | 'race' | 'alignment', value: string) => { if (characterSheet) setCharacterSheet({ ...characterSheet, [field]: value }); };
  const handleAbilityChange = (abilityName: keyof AbilityScores, newScore: number) => { if (characterSheet) setCharacterSheet({ ...characterSheet, abilityScores: { ...characterSheet.abilityScores, [abilityName]: newScore } }); };
  const handleLevelChange = (newLevel: number) => { if (characterSheet) setCharacterSheet({ ...characterSheet, level: Math.max(1, Math.min(20, newLevel)) }); };
  const handleCurrentHpChange = (newHp: number) => { if (characterSheet) setCharacterSheet({ ...characterSheet, hitPoints: { ...characterSheet.hitPoints, current: newHp } }); };
  const handleSkillProficiencyChange = (skillName: string, isProficient: boolean) => { if (characterSheet) { const newProfs = isProficient ? [...characterSheet.skillProficiencies, skillName] : characterSheet.skillProficiencies.filter(p => p !== skillName); setCharacterSheet({ ...characterSheet, skillProficiencies: newProfs }); } };
  const handleSavingThrowProficiencyChange = (abilityName: string, isProficient: boolean) => { if (characterSheet) { const newProfs = isProficient ? [...characterSheet.savingThrowProficiencies, abilityName] : characterSheet.savingThrowProficiencies.filter(p => p !== abilityName); setCharacterSheet({ ...characterSheet, savingThrowProficiencies: newProfs }); } };
  
  const handleAddItemToInventory = (item: CharacterItem) => {
    fetch('http://localhost:8080/api/character/inventory', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(item) })
      .then(res => res.json()).then(updatedInventory => setInventory(updatedInventory || [])).catch(console.error);
  };

  const handleDeleteItemFromInventory = (itemId: number) => {
    if (!confirm('Вы уверены, что хотите удалить этот предмет?')) return;
    
    // ИСПРАВЛЕНИЕ ЗДЕСЬ: Используем правильный URL
    fetch(`http://localhost:8080/api/character/inventory/item/${itemId}`, {
      method: 'DELETE',
    })
      .then(res => res.json())
      .then(updatedInventory => setInventory(updatedInventory || []))
      .catch(console.error);
  };

  const handleSaveChanges = () => {
    if (!characterSheet) return;
    fetch('http://localhost:8080/api/character', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(characterSheet) })
      .then(res => res.json()).then(updatedSheet => { setCharacterSheet(updatedSheet); alert('Персонаж сохранен!'); }).catch(console.error);
  };

  // --- РЕНДЕРИНГ ---
  if (!characterSheet) { return <div className="App-header">Загрузка данных персонажа...</div>; }

  return (
    <div className="App">
      <header className="App-header">
        <div className="header-top">
          <CharacterInfo name={characterSheet.name} characterClass={characterSheet.class} race={characterSheet.race} alignment={characterSheet.alignment} onInfoChange={handleInfoChange} />
          <div className="header-vitals">
            <Vitals level={characterSheet.level} proficiencyBonus={characterSheet.proficiencyBonus} onLevelChange={handleLevelChange} />
            <CombatStats hitPoints={characterSheet.hitPoints} armorClass={characterSheet.armorClass} initiative={characterSheet.initiative} onCurrentHpChange={handleCurrentHpChange} />
          </div>
        </div>
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
            <SavingThrowList savingThrows={characterSheet.savingThrows} proficiencies={characterSheet.savingThrowProficiencies} onProficiencyChange={handleSavingThrowProficiencyChange} />
            <button className="save-button" onClick={handleSaveChanges}>Сохранить изменения</button>
          </div>
          <SkillList skills={characterSheet.skills} skillMap={characterSheet.skillMap} proficiencies={characterSheet.skillProficiencies} onProficiencyChange={handleSkillProficiencyChange} />
          <Inventory inventory={inventory} onAddItem={handleAddItemToInventory} onDeleteItem={handleDeleteItemFromInventory} />
        </div>
      </header>
    </div>
  );
}

export default App;