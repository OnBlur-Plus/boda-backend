import { MigrationInterface, QueryRunner } from "typeorm";

export class Migration1733152550032 implements MigrationInterface {
    name = 'Migration1733152550032'

    public async up(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`CREATE TABLE \`stream\` (\`stream_key\` uuid NOT NULL, \`title\` varchar(255) NOT NULL, \`sub_title\` varchar(255) NOT NULL, \`thumbnail_url\` varchar(255) NULL, \`status\` enum ('IDLE', 'LIVE') NOT NULL, \`created_at\` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), \`updated_at\` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6), PRIMARY KEY (\`stream_key\`)) ENGINE=InnoDB`);
        await queryRunner.query(`CREATE TABLE \`user\` (\`id\` int NOT NULL AUTO_INCREMENT, \`pin\` varchar(255) NOT NULL, \`name\` varchar(255) NOT NULL, \`deviceToken\` varchar(255) NULL, UNIQUE INDEX \`IDX_798cfeabae6730aaf91c3c0463\` (\`pin\`), PRIMARY KEY (\`id\`)) ENGINE=InnoDB`);
        await queryRunner.query(`CREATE TABLE \`notification_content\` (\`id\` varchar(255) NOT NULL COMMENT 'Accident Type', \`title\` varchar(255) NOT NULL, \`body\` varchar(255) NOT NULL, \`created_at\` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), \`updated_at\` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6), PRIMARY KEY (\`id\`)) ENGINE=InnoDB`);
        await queryRunner.query(`CREATE TABLE \`accident\` (\`id\` int NOT NULL AUTO_INCREMENT, \`stream_key\` uuid NOT NULL, \`start_at\` datetime NOT NULL, \`end_at\` datetime NULL, \`type\` enum ('NON_SAFETY_VEST', 'NON_SAFETY_HELMET', 'FALL', 'USE_PHONE_WHILE_WORKING', 'SOS_REQUEST') NOT NULL, \`level\` enum ('1', '2', '3') NOT NULL, \`reason\` varchar(255) NOT NULL, \`video_url\` varchar(255) NULL, \`created_at\` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), \`updated_at\` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6), PRIMARY KEY (\`id\`)) ENGINE=InnoDB`);
        await queryRunner.query(`CREATE TABLE \`notification\` (\`id\` int NOT NULL AUTO_INCREMENT, \`accident_id\` int NOT NULL, \`notification_content_id\` varchar(255) NOT NULL, \`is_sent\` tinyint NOT NULL, \`readed_at\` datetime NULL, \`created_at\` datetime NOT NULL, \`user_id\` int NULL, PRIMARY KEY (\`id\`)) ENGINE=InnoDB`);
        await queryRunner.query(`ALTER TABLE \`accident\` ADD CONSTRAINT \`FK_e231573a6b2f4ed8c7f7983a880\` FOREIGN KEY (\`stream_key\`) REFERENCES \`stream\`(\`stream_key\`) ON DELETE NO ACTION ON UPDATE NO ACTION`);
        await queryRunner.query(`ALTER TABLE \`notification\` ADD CONSTRAINT \`FK_928b7aa1754e08e1ed7052cb9d8\` FOREIGN KEY (\`user_id\`) REFERENCES \`user\`(\`id\`) ON DELETE NO ACTION ON UPDATE NO ACTION`);
        await queryRunner.query(`ALTER TABLE \`notification\` ADD CONSTRAINT \`FK_1508b312533c71e5c822b868766\` FOREIGN KEY (\`notification_content_id\`) REFERENCES \`notification_content\`(\`id\`) ON DELETE NO ACTION ON UPDATE NO ACTION`);
        await queryRunner.query(`ALTER TABLE \`notification\` ADD CONSTRAINT \`FK_c2133fe812ee9c6849d9fac3f69\` FOREIGN KEY (\`accident_id\`) REFERENCES \`accident\`(\`id\`) ON DELETE NO ACTION ON UPDATE NO ACTION`);
    }

    public async down(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`ALTER TABLE \`notification\` DROP FOREIGN KEY \`FK_c2133fe812ee9c6849d9fac3f69\``);
        await queryRunner.query(`ALTER TABLE \`notification\` DROP FOREIGN KEY \`FK_1508b312533c71e5c822b868766\``);
        await queryRunner.query(`ALTER TABLE \`notification\` DROP FOREIGN KEY \`FK_928b7aa1754e08e1ed7052cb9d8\``);
        await queryRunner.query(`ALTER TABLE \`accident\` DROP FOREIGN KEY \`FK_e231573a6b2f4ed8c7f7983a880\``);
        await queryRunner.query(`DROP TABLE \`notification\``);
        await queryRunner.query(`DROP TABLE \`accident\``);
        await queryRunner.query(`DROP TABLE \`notification_content\``);
        await queryRunner.query(`DROP INDEX \`IDX_798cfeabae6730aaf91c3c0463\` ON \`user\``);
        await queryRunner.query(`DROP TABLE \`user\``);
        await queryRunner.query(`DROP TABLE \`stream\``);
    }

}
