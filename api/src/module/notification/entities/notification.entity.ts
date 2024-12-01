import {
  Column,
  CreateDateColumn,
  Entity,
  JoinColumn,
  ManyToOne,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm';
import { User } from 'src/module/auth/entities/user.entity';
import { NotificationContent } from './notification-content.entity';
import { Accident } from 'src/module/accident/entities/accident.entity';

@Entity()
export class Notification {
  @PrimaryGeneratedColumn()
  id: number;

  @JoinColumn({ name: 'user_id' })
  @ManyToOne(() => User, { cascade: true })
  user: User;

  @Column({ name: 'accident_id' })
  accidentId: number;

  @Column({ name: 'notification_content_id' })
  notificationContentId: string;

  @Column({ name: 'is_sent' })
  isSent: boolean;

  @Column({ name: 'readed_at', nullable: true })
  readedAt: Date | null;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;

  @JoinColumn({ name: 'notification_content_id' })
  @ManyToOne(() => NotificationContent, { cascade: true })
  notificationContent: NotificationContent;

  @JoinColumn({ name: 'accident_id' })
  @ManyToOne(() => Accident, { cascade: true })
  accident: Accident;
}
