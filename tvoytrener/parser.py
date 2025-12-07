import requests
from bs4 import BeautifulSoup
import re

def parse_muscles_from_tvoytrener():
    url = "https://tvoytrener.com/www/index.html"
    
    try:
        # Отправляем запрос к странице
        response = requests.get(url)
        response.raise_for_status()
        
        # Создаем объект BeautifulSoup
        soup = BeautifulSoup(response.content, 'html.parser')
        
        muscles = set()
        
        # Ищем все ссылки на упражнения (они обычно содержат названия мышц)
        exercise_links = soup.find_all('a', href=re.compile(r'yprajnenia/'))
        
        for link in exercise_links:
            text = link.get_text().strip()
            if text:
                # Добавляем текст ссылки как потенциальное название упражнения/мышцы
                muscles.add(text)
        
        # Также ищем в других разделах
        list_items = soup.find_all('li')
        for item in list_items:
            text = item.get_text().strip()
            if text and any(keyword in text.lower() for keyword in ['мышц', 'упражнен', 'жим', 'тяга', 'разгибан', 'сгибан']):
                muscles.add(text.split('-')[0].strip())
        
        # Преобразуем в отсортированный список
        muscles_list = sorted(list(muscles))
        
        return muscles_list
        
    except requests.RequestException as e:
        print(f"Ошибка при запросе: {e}")
        return []
    except Exception as e:
        print(f"Ошибка при парсинге: {e}")
        return []

def get_muscles_from_exercises():
    """Альтернативный метод - парсим конкретные упражнения"""
    
    # Список основных групп мышц в бодибилдинге
    main_muscle_groups = [
        'Грудь', 'Спина', 'Плечи', 'Бицепс', 'Трицепс', 'Предплечья',
        'Квадрицепсы', 'Бицепсы бедер', 'Ягодицы', 'Икры', 'Пресс',
        'Трапеции', 'Широчайшие', 'Дельтовидные', 'Поясница'
    ]
    
    return sorted(main_muscle_groups)

if __name__ == "__main__":
    print("Парсим мышцы с сайта Твой Тренер...")
    
    # Парсим с сайта
    parsed_muscles = parse_muscles_from_tvoytrener()
    
    print("\n=== Найденные упражнения и мышцы ===")
    for i, muscle in enumerate(parsed_muscles, 1):
        print(f"{i}. {muscle}")
    
    print(f"\nВсего найдено: {len(parsed_muscles)}")
    
    # Также выводим стандартный список мышц
    print("\n=== Основные группы мышц ===")
    standard_muscles = get_muscles_from_exercises()
    for i, muscle in enumerate(standard_muscles, 1):
        print(f"{i}. {muscle}")