# correlation_matrix.R
#
# Cross-instrument correlation analysis over ingested market data.
# Builds a return-correlation matrix across a basket of symbols so the
# risk desk can spot concentration risk before it shows up in a stress
# test.

suppressWarnings(suppressMessages({
  library(stats)
}))

load_basket_quotes <- function(path) {
  quotes <- read.csv(path, stringsAsFactors = FALSE)
  quotes$ingested_at <- as.POSIXct(quotes$ingested_at_ms / 1000, origin = "1970-01-01", tz = "UTC")
  quotes[order(quotes$symbol, quotes$ingested_at), ]
}

pivot_prices_by_symbol <- function(quotes) {
  symbols <- unique(quotes$symbol)
  series_list <- lapply(symbols, function(sym) {
    quotes[quotes$symbol == sym, c("ingested_at", "last_price")]
  })
  names(series_list) <- symbols
  series_list
}

align_and_compute_returns <- function(series_list) {
  min_length <- min(vapply(series_list, function(s) nrow(s), integer(1)))
  returns_by_symbol <- lapply(series_list, function(s) {
    prices <- tail(s$last_price, min_length)
    diff(log(prices))
  })
  as.data.frame(returns_by_symbol)
}

build_correlation_matrix <- function(returns_df) {
  cor(returns_df, use = "pairwise.complete.obs")
}

find_highest_pairwise_correlation <- function(correlation_matrix) {
  symbols <- rownames(correlation_matrix)
  best_pair <- NULL
  best_value <- -Inf

  for (i in seq_along(symbols)) {
    for (j in seq_along(symbols)) {
      if (i >= j) {
        next
      }
      value <- correlation_matrix[i, j]
      if (!is.na(value) && value > best_value) {
        best_value <- value
        best_pair <- c(symbols[i], symbols[j])
      }
    }
  }

  list(pair = best_pair, correlation = best_value)
}

summarize_basket_correlation <- function(path) {
  quotes <- load_basket_quotes(path)
  series_list <- pivot_prices_by_symbol(quotes)
  returns_df <- align_and_compute_returns(series_list)
  correlation_matrix <- build_correlation_matrix(returns_df)

  list(
    symbols = colnames(correlation_matrix),
    correlation_matrix = correlation_matrix,
    highest_pair = find_highest_pairwise_correlation(correlation_matrix)
  )
}

main <- function() {
  args <- commandArgs(trailingOnly = TRUE)
  if (length(args) < 1) {
    stop("Usage: Rscript correlation_matrix.R <path-to-basket-quotes-csv>")
  }

  summary_result <- summarize_basket_correlation(args[1])
  cat(sprintf("Symbols in basket: %s\n", paste(summary_result$symbols, collapse = ", ")))
  cat("Correlation matrix:\n")
  print(round(summary_result$correlation_matrix, 3))

  if (!is.null(summary_result$highest_pair$pair)) {
    cat(sprintf(
      "\nHighest pairwise correlation: %s / %s (%.3f)\n",
      summary_result$highest_pair$pair[1],
      summary_result$highest_pair$pair[2],
      summary_result$highest_pair$correlation
    ))
  }
}

if (!interactive()) {
  main()
}
