CREATE TEXT SEARCH DICTIONARY russian_hunspell (
    TEMPLATE = ispell,
    DictFile = ru_RU,      
    AffFile = ru_RU,       
    StopWords = russian    
);

CREATE TEXT SEARCH CONFIGURATION ru_hunspell (COPY = russian);

ALTER TEXT SEARCH CONFIGURATION ru_hunspell
ALTER MAPPING FOR hword, hword_part, word
WITH russian_hunspell, russian_stem;

ALTER TEXT SEARCH CONFIGURATION ru_hunspell RENAME TO ru;
