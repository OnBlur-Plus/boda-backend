import { DataSource, Repository } from 'typeorm';
import FirebaseAdmin from 'firebase-admin';
import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { RegisterDeviceTokenDto } from './dto/register-device-token.dto';
import { User } from '../auth/entities/user.entity';
import { Notification } from './entities/notification.entity';
import { NotificationContent } from './entities/notification-content.entity';
import { FirebaseConfigService } from '../config/firebase-config.service';

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
    return await this.notificationRepository.find({
      where: { user: { id: userId } },
      skip: (pageNum - 1) * pageSize,
      take: pageSize,
    });
  }

  async registerDeviceToken(userId: number, { deviceToken }: RegisterDeviceTokenDto) {
    return await this.userRepository.update({ id: userId }, { deviceToken });
  }

  async sendNotificationToAllUsers(title: string, body: string, accidentId: number) {
    return await this.dataSource.transaction(async (manager) => {
      const [users, notificationContent] = await Promise.all([
        manager.find(User),
        manager.save(NotificationContent, { title, body, accident: { id: accidentId } }),
      ]);

      const results = await FirebaseAdmin.messaging().sendEachForMulticast({
        tokens: users.map((user) => user.deviceToken),
        notification: { title, body },
        data: { accidentId: accidentId.toString() },
      });

      return await manager.save(
        Notification,
        users.map((user, index) => ({
          user: { id: user.id },
          notificationContent: { id: notificationContent.id },
          isSent: results.responses[index].success,
        })),
      );
    });
  }

  async readAllNotifications(userId: number, date: Date) {
    const { affected } = await this.notificationRepository.update(
      { user: { id: userId }, readedAt: null },
      { readedAt: date },
    );

    return affected;
  }
}
