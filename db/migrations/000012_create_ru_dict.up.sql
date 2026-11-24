-- Создаём словарь на основе Hunspell файлов
CREATE TEXT SEARCH DICTIONARY russian_hunspell (
    TEMPLATE = ispell,
    DictFile = ru_RU,      -- без расширения .dict
    AffFile = ru_RU,       -- без расширения .affix
    StopWords = russian    -- можно использовать стандартный стоп-лист
);

-- Создаём конфигурацию (копируем стандартную russian)
CREATE TEXT SEARCH CONFIGURATION ru_hunspell (COPY = russian);

-- Привязываем наш словарь к типам токенов
ALTER TEXT SEARCH CONFIGURATION ru_hunspell
ALTER MAPPING FOR hword, hword_part, word
WITH russian_hunspell, russian_stem;  -- russian_stem как fallback

ALTER TEXT SEARCH CONFIGURATION ru_hunspell RENAME TO ru;
