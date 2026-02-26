-- Удаление модуля «Умный дом на Pico W» и всех его уроков
DELETE FROM user_checklist_progress WHERE checklist_item_id IN (SELECT id FROM checklist_items WHERE lesson_id IN (SELECT id FROM lessons WHERE module_id = 'a0000003-0000-0000-0000-000000000003'));
DELETE FROM checklist_items WHERE lesson_id IN (SELECT id FROM lessons WHERE module_id = 'a0000003-0000-0000-0000-000000000003');
DELETE FROM lesson_tags WHERE lesson_id IN (SELECT id FROM lessons WHERE module_id = 'a0000003-0000-0000-0000-000000000003');
DELETE FROM lesson_materials WHERE lesson_id IN (SELECT id FROM lessons WHERE module_id = 'a0000003-0000-0000-0000-000000000003');
DELETE FROM lesson_steps WHERE lesson_id IN (SELECT id FROM lessons WHERE module_id = 'a0000003-0000-0000-0000-000000000003');
DELETE FROM user_lesson_progress WHERE lesson_id IN (SELECT id FROM lessons WHERE module_id = 'a0000003-0000-0000-0000-000000000003');
DELETE FROM lessons WHERE module_id = 'a0000003-0000-0000-0000-000000000003';
DELETE FROM modules WHERE id = 'a0000003-0000-0000-0000-000000000003';
