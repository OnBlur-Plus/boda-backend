import {
  Column,
  CreateDateColumn,
  Entity,
  JoinColumn,
  ManyToOne,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm';
import { Accident } from 'src/module/accident/entities/accident.entity';

@Entity()
export class NotificationContent {
  @PrimaryGeneratedColumn()
  id: number;

  @Column()
  title: string;

  @Column()
  body: string;

  @Column({ name: 'accident_id' })
  accidentId: number;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;

  @JoinColumn({ name: 'accident_id' })
  @ManyToOne(() => Accident, { cascade: true })
  accident: Accident;
}
