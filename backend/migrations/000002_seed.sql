-- Catégories principales (8)
INSERT INTO categories (name_ar, name_fr, icon_name, display_order) VALUES
    ('عقارات',     'Immobilier',    'home',         1),
    ('سيارات',     'Voitures',      'car',          2),
    ('هواتف',      'Téléphones',    'phone',        3),
    ('إلكترونيات', 'Électronique',  'monitor',      4),
    ('ساعات',      'Montres',       'watch',        5),
    ('دراجات',     'Motos',         'bike',         6),
    ('حيوانات',    'Animaux',       'paw',          7),
    ('أثاث',       'Meubles',       'sofa',         8);

-- Localisations (Mauritanie)
INSERT INTO locations (city_name, area_name) VALUES
    ('Nouakchott',  'Tevragh Zeina'),
    ('Nouakchott',  'Ksar'),
    ('Nouakchott',  'Arafat'),
    ('Nouakchott',  'El Mina'),
    ('Nouakchott',  'Dar Naim'),
    ('Nouakchott',  'Teyaret'),
    ('Nouakchott',  'Toujounine'),
    ('Nouakchott',  'Sebkha'),
    ('Nouadhibou',  'Cansado'),
    ('Nouadhibou',  'Centre ville');
