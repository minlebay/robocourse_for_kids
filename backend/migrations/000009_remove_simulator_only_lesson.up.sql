-- Убираем отдельный урок «Тренажёр без платы» — тренажёр остаётся только в уроке «Первая программа»
DELETE FROM lesson_materials WHERE lesson_id = 'b0000007-0000-0000-0000-000000000001';
DELETE FROM lesson_steps WHERE lesson_id = 'b0000007-0000-0000-0000-000000000001';
DELETE FROM lesson_tags WHERE lesson_id = 'b0000007-0000-0000-0000-000000000001';
DELETE FROM lessons WHERE id = 'b0000007-0000-0000-0000-000000000001';
