import { DataSource, IsNull, Repository } from 'typeorm';
import FirebaseAdmin from 'firebase-admin';
import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { RegisterDeviceTokenDto } from './dto/register-device-token.dto';
import { User } from '../auth/entities/user.entity';
import { Notification } from './entities/notification.entity';
import { NotificationContent } from './entities/notification-content.entity';
import { FirebaseConfigService } from '../config/firebase-config.service';
import { Accident, AccidentLevel } from '../accident/entities/accident.entity';

@Injectable()
export class NotificationService {
  constructor(
    @InjectRepository(User) private readonly userRepository: Repository<User>,
    @InjectRepository(Notification) private readonly notificationRepository: Repository<Notification>,
    private readonly dataSource: DataSource,
    firebaseConfigService: FirebaseConfigService,
  ) {
    FirebaseAdmin.initializeApp({ credential: FirebaseAdmin.credential.cert(firebaseConfigService) });
  }

  async getAllNotifications(userId: number, pageNum: number, pageSize: number) {
    const [notifications, count] = await this.notificationRepository.findAndCount({
      where: { user: { id: userId } },
      skip: (pageNum - 1) * pageSize,
      take: pageSize,
      order: { createdAt: 'DESC' },
      relations: ['notificationContent'],
    });

    return { pageNum, pageSize, hasNext: count > pageNum * pageSize, notifications };
  }

  async registerDeviceToken(userId: number, { deviceToken }: RegisterDeviceTokenDto) {
    return await this.userRepository.update({ id: userId }, { deviceToken });
  }

  async sendNotificationToAllUsers(accident: Accident) {
    return await this.dataSource.transaction(async (manager) => {
      const [users, notificationContent] = await Promise.all([
        manager.find(User),
        manager.findOne(NotificationContent, { where: { id: accident.type } }),
      ]);

      const results = await FirebaseAdmin.messaging().sendEachForMulticast({
        tokens: users.map((user) => user.deviceToken),
        notification: { title: notificationContent.title, body: notificationContent.body },
        android: { priority: accident.level === AccidentLevel.HIGH ? 'high' : 'normal' },
        data: { accidentId: accident.id.toString() },
      });

      const createdAt = new Date();

      return await manager.save(
        Notification,
        users.map((user, index) => ({
          user: { id: user.id },
          accident: { id: accident.id },
          notificationContent: { id: notificationContent.id },
          isSent: results.responses[index].success,
          createdAt,
        })),
      );
    });
  }

  async readAllNotifications(userId: number, date: Date) {
    const { affected } = await this.notificationRepository.update(
      { user: { id: userId }, readedAt: IsNull() },
      { readedAt: date },
    );

    return affected;
  }
}
