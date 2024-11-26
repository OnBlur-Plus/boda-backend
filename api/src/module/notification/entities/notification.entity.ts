import { Column, CreateDateColumn, Entity, JoinColumn, ManyToOne, PrimaryGeneratedColumn } from 'typeorm';
import { Accident } from 'src/module/accident/entities/accident.entities';
import { User } from 'src/module/auth/entities/user.entity';

@Entity()
export class Notification {
  @PrimaryGeneratedColumn()
  id: number;

  @JoinColumn({ name: 'user_id' })
  @ManyToOne(() => User)
  user: User;

  @JoinColumn({ name: 'accident_id' })
  @ManyToOne(() => Accident)
  accident: Accident;

  @Column({ name: 'readed_at', nullable: true })
  readedAt: Date;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;
}
