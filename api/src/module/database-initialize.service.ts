import { Injectable, OnApplicationBootstrap } from '@nestjs/common';
import { DataSource } from 'typeorm';

@Injectable()
export class DatabaseInitializeService implements OnApplicationBootstrap {
  constructor(private readonly dataSource: DataSource) {}

  async onApplicationBootstrap() {
    console.log('Running SQL scripts...');
    await this.dataSource.query(`
      INSERT INTO db.notification_content (id,title,body,created_at,updated_at) VALUES
	 ('FALL','재해 발생 알림','낙상 사고가 발생했어요.','2024-12-01 11:28:01.565251','2024-12-01 11:28:01.565251'),
	 ('NON_SAFETY_HELMET','재해 경고 알림','안전모를 착용하지 않은 근로자를 발견했어요.','2024-12-01 11:28:01.558575','2024-12-01 11:28:01.558575'),
	 ('NON_SAFETY_VEST','재해 경고 알림','안전 조끼를 착용하지 않은 근로자를 발견했어요.','2024-12-01 11:27:17.967078','2024-12-01 11:27:17.967078'),
	 ('SOS_REQUEST','비상 상황 알림','구조 요청을 보내는 근로자를 발견했어요.','2024-12-01 11:28:01.567892','2024-12-01 11:28:01.567892'),
	 ('USE_PHONE_WHILE_WORKING','재해 경고 알림','보행 중 휴대폰을 사용하는 근로자를 발견했어요.','2024-12-01 11:28:01.562606','2024-12-01 11:28:01.562606');

    `);
    console.log('SQL scripts executed successfully.');
  }
}
