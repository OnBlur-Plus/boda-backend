import { AccidentLevel, AccidentType } from 'src/module/accident/entities/accident.entity';

export const ACCIDENT_METADATA: Record<AccidentType, { reason: string; level: AccidentLevel }> = {
  [AccidentType.NON_SAFETY_HELMET]: { reason: '안전모 미착용', level: AccidentLevel.LOW },
  [AccidentType.NON_SAFETY_VEST]: { reason: '안전 조끼 미착용', level: AccidentLevel.LOW },
  [AccidentType.USE_PHONE_WHILE_WORKING]: { reason: '보행 중 휴대폰 사용', level: AccidentLevel.LOW },
  [AccidentType.FALL]: { reason: '낙상 사고 발생', level: AccidentLevel.MEDIUM },
  [AccidentType.SOS_REQUEST]: { reason: '구조 요청', level: AccidentLevel.HIGH },
};
