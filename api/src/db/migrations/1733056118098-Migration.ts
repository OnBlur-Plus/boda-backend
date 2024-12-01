import { MigrationInterface, QueryRunner } from "typeorm";

export class Migration1733056118098 implements MigrationInterface {
    name = 'Migration1733056118098'

    public async up(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`ALTER TABLE \`stream\` CHANGE \`sub_title\` \`sub_title\` varchar(255) NOT NULL`);
        await queryRunner.query(`ALTER TABLE \`notification\` DROP FOREIGN KEY \`FK_1508b312533c71e5c822b868766\``);
        await queryRunner.query(`ALTER TABLE \`notification_content\` CHANGE \`id\` \`id\` int NOT NULL`);
        await queryRunner.query(`ALTER TABLE \`notification_content\` DROP PRIMARY KEY`);
        await queryRunner.query(`ALTER TABLE \`notification_content\` DROP COLUMN \`id\``);
        await queryRunner.query(`ALTER TABLE \`notification_content\` ADD \`id\` varchar(255) NOT NULL PRIMARY KEY COMMENT 'Accident Type'`);
        await queryRunner.query(`ALTER TABLE \`accident\` CHANGE \`reason\` \`reason\` varchar(255) NOT NULL`);
        await queryRunner.query(`ALTER TABLE \`notification\` DROP COLUMN \`notification_content_id\``);
        await queryRunner.query(`ALTER TABLE \`notification\` ADD \`notification_content_id\` varchar(255) NOT NULL`);
        await queryRunner.query(`ALTER TABLE \`notification\` CHANGE \`created_at\` \`created_at\` datetime NOT NULL`);
        await queryRunner.query(`ALTER TABLE \`notification\` ADD CONSTRAINT \`FK_1508b312533c71e5c822b868766\` FOREIGN KEY (\`notification_content_id\`) REFERENCES \`notification_content\`(\`id\`) ON DELETE NO ACTION ON UPDATE NO ACTION`);
        await queryRunner.query(`ALTER TABLE \`notification\` ADD CONSTRAINT \`FK_c2133fe812ee9c6849d9fac3f69\` FOREIGN KEY (\`accident_id\`) REFERENCES \`accident\`(\`id\`) ON DELETE NO ACTION ON UPDATE NO ACTION`);
    }

    public async down(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`ALTER TABLE \`notification\` DROP FOREIGN KEY \`FK_c2133fe812ee9c6849d9fac3f69\``);
        await queryRunner.query(`ALTER TABLE \`notification\` DROP FOREIGN KEY \`FK_1508b312533c71e5c822b868766\``);
        await queryRunner.query(`ALTER TABLE \`notification\` CHANGE \`created_at\` \`created_at\` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)`);
        await queryRunner.query(`ALTER TABLE \`notification\` DROP COLUMN \`notification_content_id\``);
        await queryRunner.query(`ALTER TABLE \`notification\` ADD \`notification_content_id\` int NULL`);
        await queryRunner.query(`ALTER TABLE \`accident\` CHANGE \`reason\` \`reason\` varchar(255) NULL`);
        await queryRunner.query(`ALTER TABLE \`notification_content\` DROP COLUMN \`id\``);
        await queryRunner.query(`ALTER TABLE \`notification_content\` ADD \`id\` int NOT NULL AUTO_INCREMENT`);
        await queryRunner.query(`ALTER TABLE \`notification_content\` ADD PRIMARY KEY (\`id\`)`);
        await queryRunner.query(`ALTER TABLE \`notification_content\` CHANGE \`id\` \`id\` int NOT NULL AUTO_INCREMENT`);
        await queryRunner.query(`ALTER TABLE \`notification\` ADD CONSTRAINT \`FK_1508b312533c71e5c822b868766\` FOREIGN KEY (\`notification_content_id\`) REFERENCES \`notification_content\`(\`id\`) ON DELETE NO ACTION ON UPDATE NO ACTION`);
        await queryRunner.query(`ALTER TABLE \`stream\` CHANGE \`sub_title\` \`sub_title\` varchar(255) NULL`);
    }

}
