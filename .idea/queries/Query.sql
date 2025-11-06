
DELIMITER $$

DROP PROCEDURE generate_data;

CREATE PROCEDURE generate_data()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 124000 DO
            INSERT INTO `symbol-service`.symbols (project_id, uid, label, class_name, component_target, version, created_at, updated_at)
                            VALUES (1,
                                    CONCAT('uid_',ROUND(RAND()*200700,2),ROUND(RAND()*200900,2)),
                                    CONCAT('label_',ROUND(RAND()*200600,2),ROUND(RAND()*200900,2)),
                                    CONCAT('class_name_',ROUND(RAND()*20060,2),ROUND(RAND()*290000,2)),
                                    CONCAT('compoenent_taget',ROUND(RAND()*200600,2),ROUND(RAND()*200900,2)),
                                    i,
                                    NOW(),
                                    NOW()
                                   );

            SET i = i + 1;
        END WHILE;
END$$
DELIMITER ;



CALL generate_data();


SELECT count(*) FROM `symbol-service`.symbols;



    La Multi Ani Tinere!
    Zboruri inalte si realizari pe toate planurile!
    Ai grija de tine si de cei dragi tie!


