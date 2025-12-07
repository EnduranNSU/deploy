import json
import re
from typing import Dict, List, Any

class ExerciseParser:
    def __init__(self):
        self.muscle_groups = {}
        
    def parse_javascript_code(self, js_code: str, muscle_mapping: Dict[str, str]) -> Dict[str, Any]:
        """
        Парсит JavaScript код и извлекает упражнения по группам мышц
        """
        results = {}
        
        # Ищем все переменные с ключами мышц
        for muscle_key, muscle_name in muscle_mapping.items():
            # Создаем паттерн для поиска всей строки до точки с запятой
            pattern = rf'{muscle_key}2\s*=\s*document\.getElementById\([^)]+\)\.innerHTML\s*=\s*([^;]+);'
            match = re.search(pattern, js_code, re.DOTALL)
            
            if match:
                html_assignment = match.group(1)
                # Обрабатываем конкатенацию строк
                html_content = self._process_string_concat(html_assignment)
                exercises = self._get_exercises(html_content)
                results[muscle_name] = {
                    'key': muscle_key,
                    'exercises': exercises
                }
                print(f"Найдены упражнения для: {muscle_name} ({muscle_key})")
            else:
                print(f"Не найдены упражнения для: {muscle_name} ({muscle_key})")
            
        return results
    
    def _process_string_concat(self, assignment: str) -> str:
        """
        Обрабатывает конкатенацию строк JavaScript и объединяет их в одну строку
        """
        # Убираем лишние пробелы и переносы
        assignment = assignment.strip()
        
        # Если это просто строка в кавычках
        if assignment.startswith('"') and assignment.endswith('"'):
            return assignment[1:-1]
        
        # Обрабатываем конкатенацию вида "string" + variable + "string"
        parts = re.split(r'\s*\+\s*', assignment)
        result_parts = []
        
        for part in parts:
            part = part.strip()
            # Если часть - строка в кавычках
            if part.startswith('"') and part.endswith('"'):
                result_parts.append(part[1:-1])
            # Если часть - переменная (например zu)
            elif part == 'zu':
                result_parts.append('')  # заменяем переменную на пустую строку
            else:
                result_parts.append(part)
        
        return ''.join(result_parts)
    
    def _get_exercises(self, content: str) -> Dict[str, str]:
        all_exercises = {}
        pattern = r"<li class='[^']*'\\\"\+zu\+\\\"([^<]+)'>([^<]+)<\/a><\/li>"
        matches = re.findall(pattern, content)
        for html_file, exercise_name in matches:
            all_exercises[exercise_name] = html_file
        
        return all_exercises

    
    def parse_from_file(self, js_file_path: str, mapping_file_path: str) -> Dict[str, Any]:
        """
        Читает данные из файлов и парсит JavaScript код
        """
        # Читаем mapping из JSON файла
        with open(mapping_file_path, 'r', encoding='utf-8') as f:
            muscle_mapping = json.load(f)
        
        # Читаем JavaScript код
        with open(js_file_path, 'r', encoding='utf-8') as f:
            js_code = f.read()
        
        return self.parse_javascript_code(js_code, muscle_mapping)
    
    def save_to_json(self, data: Dict[str, Any], output_file: str):
        """
        Сохраняет результаты в JSON файл
        """
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

# Пример использования
if __name__ == "__main__":
    parser = ExerciseParser()
    
    # Парсим данные
    try:
        results = parser.parse_from_file('poisk.js', 'muscles.json')
        
        # Сохраняем результаты
        parser.save_to_json(results, 'parsed_exercises.json')
        
        print(f"\nНайдено групп мышц: {len(results)}")
        print("Результаты сохранены в файл 'parsed_exercises.json'")
        
    except FileNotFoundError as e:
        print(f"Ошибка: Файл не найден - {e}")
    except json.JSONDecodeError as e:
        print(f"Ошибка: Неверный формат JSON файла - {e}")
    except Exception as e:
        print(f"Произошла ошибка: {e}")