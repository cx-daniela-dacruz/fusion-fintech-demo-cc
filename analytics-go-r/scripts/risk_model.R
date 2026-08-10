# risk_model.R
#
# Statistical risk modeling over ingested market data. Reads normalized
# daily quotes exported by the Go ingestion service and computes a
# small set of risk metrics (log returns, annualized volatility,
# historical VaR, and a simple mean-reversion fit) used by the risk
# desk's daily report.

suppressWarnings(suppressMessages({
  library(stats)
}))

load_quotes <- function(path) {
  quotes <- read.csv(path, stringsAsFactors = FALSE)
  quotes$ingested_at <- as.POSIXct(quotes$ingested_at_ms / 1000, origin = "1970-01-01", tz = "UTC")
  quotes[order(quotes$ingested_at), ]
}

compute_log_returns <- function(prices) {
  diff(log(prices))
}

annualized_volatility <- function(returns, periods_per_year = 252) {
  sd(returns, na.rm = TRUE) * sqrt(periods_per_year)
}

historical_var <- function(returns, confidence = 0.95) {
  quantile(returns, probs = 1 - confidence, na.rm = TRUE)
}

fit_mean_reversion <- function(prices) {
  n <- length(prices)
  lagged <- prices[1:(n - 1)]
  current <- prices[2:n]
  lm(current ~ lagged)
}

summarize_risk <- function(path) {
  quotes <- load_quotes(path)
  returns <- compute_log_returns(quotes$last_price)

  list(
    symbol = unique(quotes$symbol),
    observations = nrow(quotes),
    annualized_vol = annualized_volatility(returns),
    var_95 = historical_var(returns, 0.95),
    var_99 = historical_var(returns, 0.99),
    mean_reversion_model = fit_mean_reversion(quotes$last_price)
  )
}

main <- function() {
  args <- commandArgs(trailingOnly = TRUE)
  if (length(args) < 1) {
    stop("Usage: Rscript risk_model.R <path-to-quotes-csv>")
  }

  risk_summary <- summarize_risk(args[1])
  cat(sprintf("Symbol(s): %s\n", paste(risk_summary$symbol, collapse = ", ")))
  cat(sprintf("Observations: %d\n", risk_summary$observations))
  cat(sprintf("Annualized volatility: %.4f\n", risk_summary$annualized_vol))
  cat(sprintf("95%% historical VaR: %.4f\n", risk_summary$var_95))
  cat(sprintf("99%% historical VaR: %.4f\n", risk_summary$var_99))
}

if (!interactive()) {
  main()
}
