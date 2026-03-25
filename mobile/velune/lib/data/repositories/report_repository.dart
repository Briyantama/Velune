import '../../core/http/dio_client.dart';
import '../../core/json/normalized_deserialize.dart';
import '../../domain/models/monthly_report.dart';

abstract class ReportRepository {
  Future<MonthlyReport> monthly({
    required int year,
    required int month,
    required String currency,
  });
}

class DioReportRepository implements ReportRepository {
  final DioClient _client;

  DioReportRepository({required DioClient client}) : _client = client;

  @override
  Future<MonthlyReport> monthly({
    required int year,
    required int month,
    required String currency,
  }) async {
    final resp = await _client.dio.get(
      '/reports/monthly',
      queryParameters: {
        'year': year,
        'month': month,
        'currency': currency,
      },
    );

    return fromJsonNormalized<MonthlyReport>(
      resp.data,
      fromJson: MonthlyReport.fromJson,
    );
  }
}

